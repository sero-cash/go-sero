// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package keystore

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cespare/cp"
	"github.com/davecgh/go-spew/spew"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/common"
)

var (
	cachetestDir, _   = filepath.Abs(filepath.Join("testdata", "keystore"))
	cachetestAccounts = []accounts.Account{
		{
			Address: common.Base58ToAddress("4dDHHzaHCr5eKgkEhtfCr9qQDBisDjJfka4zms94D51QKffCm1F1YzUdTofxafYYaNijTrNfWbsFZwiepmHkY7Pz"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(cachetestDir, "UTC--2018-09-11T10-50-15.518299056Z--4dDHHzaHCr5eKgkEhtfCr9qQDBisDjJfka4zms94D51QKffCm1F1YzUdTofxafYYaNijTrNfWbsFZwiepmHkY7Pz")},
		},
		{
			Address: common.Base58ToAddress("TKtF23g7WfnTFkSbskERmaMiMxUCCgPfG4rMRGEmrRCGrUFnHULC4XA6WEF8HxGF3JbQc4f8wj8Hg9T7yv91KHs"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(cachetestDir, "aaa")},
		},
		{
			Address: common.Base58ToAddress("2DnJDxmGkeFvvxxMP6ZUwJNQx5uGEtRfY6ahpQBkzQruouvL7sCwbgTHTFN3hernDpM2bLbUQdZds3PyUBHqaAzA"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(cachetestDir, "zzz")},
		},
	}
)

func TestWatchNewFile(t *testing.T) {
	t.Parallel()

	dir, ks := tmpKeyStore(t)
	defer os.RemoveAll(dir)

	// Ensure the watcher is started before adding any files.
	ks.Accounts()
	time.Sleep(1000 * time.Millisecond)

	// Move in the files.
	wantAccounts := make([]accounts.Account, len(cachetestAccounts))
	for i := range cachetestAccounts {
		wantAccounts[i] = accounts.Account{
			Address: cachetestAccounts[i].Address,
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, filepath.Base(cachetestAccounts[i].URL.Path))},
		}
		if err := cp.CopyFile(wantAccounts[i].URL.Path, cachetestAccounts[i].URL.Path); err != nil {
			t.Fatal(err)
		}
	}

	// ks should see the accounts.
	var list []accounts.Account
	for d := 200 * time.Millisecond; d < 5*time.Second; d *= 2 {
		list = ks.Accounts()
		if reflect.DeepEqual(list, wantAccounts) {
			// ks should have also received change notifications
			select {
			case <-ks.changes:
			default:
				t.Fatalf("wasn't notified of new accounts")
			}
			return
		}
		time.Sleep(d)
	}
	t.Errorf("got %s, want %s", spew.Sdump(list), spew.Sdump(wantAccounts))
}

func TestWatchNoDir(t *testing.T) {
	t.Parallel()

	// Create ks but not the directory that it watches.
	rand.Seed(time.Now().UnixNano())
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("sero-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))
	ks := NewKeyStore(dir, LightScryptN, LightScryptP)

	list := ks.Accounts()
	if len(list) > 0 {
		t.Error("initial account list not empty:", list)
	}
	time.Sleep(100 * time.Millisecond)

	// Create the directory and copy a key file into it.
	os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "aaa")
	if err := cp.CopyFile(file, cachetestAccounts[0].URL.Path); err != nil {
		t.Fatal(err)
	}

	// ks should see the account.
	wantAccounts := []accounts.Account{cachetestAccounts[0]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	for d := 200 * time.Millisecond; d < 8*time.Second; d *= 2 {
		list = ks.Accounts()
		if reflect.DeepEqual(list, wantAccounts) {
			// ks should have also received change notifications
			select {
			case <-ks.changes:
			default:
				t.Fatalf("wasn't notified of new accounts")
			}
			return
		}
		time.Sleep(d)
	}
	t.Errorf("\ngot  %v\nwant %v", list, wantAccounts)
}

func TestCacheInitialReload(t *testing.T) {
	cache, _ := newAccountCache(cachetestDir)
	accounts := cache.accounts()
	if !reflect.DeepEqual(accounts, cachetestAccounts) {
		t.Fatalf("got initial accounts: %swant %s", spew.Sdump(accounts), spew.Sdump(cachetestAccounts))
	}
}

func TestCacheAddDeleteOrder(t *testing.T) {
	cache, _ := newAccountCache("testdata/no-such-dir")
	cache.watcher.running = true // prevent unexpected reloads

	accs := []accounts.Account{
		{
			Address: common.Base58ToAddress("4Ty8ZMWsP6EUP7jDDmN4tmxgTa4LfWpwbaDQPBxNRNibfhhWC6RhGfLwDnUu5FHDqxnfKjHm3NhFaBm1ZP5cB664"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "-309830980"},
		},
		{
			Address: common.Base58ToAddress("5TBbwPWNZo1cZ54vwrnM32GLgRRLExSfjwxj1r4DiL2don3bZj3jDHaweCjsidBZEMarLJCcTZDBRAJQpMNgQxsQ"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "ggg"},
		},
		{
			Address: common.Base58ToAddress("2V6DVW81tuaoC1USymhvYAxENvTuRo9o9hMWqAcfRb9XJczFH2b8zJnyondpQytMSEaFQn7BfhHPWGvxfsVPoGPL"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "zzzzzz-the-very-last-one.keyXXX"},
		},
		{
			Address: common.Base58ToAddress("4dDHHzaHCr5eKgkEhtfCr9qQDBisDjJfka4zms94D51QKffCm1F1YzUdTofxafYYaNijTrNfWbsFZwiepmHkY7Pz"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "SOMETHING.key"},
		},
		{
			Address: common.Base58ToAddress("7ef5a6135f1fd6a02593eedc869c6d41d934aef8"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "UTC--2018-09-11T10-50-15.518299056Z--4dDHHzaHCr5eKgkEhtfCr9qQDBisDjJfka4zms94D51QKffCm1F1YzUdTofxafYYaNijTrNfWbsFZwiepmHkY7Pz"},
		},
		{
			Address: common.Base58ToAddress("TKtF23g7WfnTFkSbskERmaMiMxUCCgPfG4rMRGEmrRCGrUFnHULC4XA6WEF8HxGF3JbQc4f8wj8Hg9T7yv91KHs"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "aaa"},
		},
		{
			Address: common.Base58ToAddress("2DnJDxmGkeFvvxxMP6ZUwJNQx5uGEtRfY6ahpQBkzQruouvL7sCwbgTHTFN3hernDpM2bLbUQdZds3PyUBHqaAzA"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: "zzz"},
		},
	}
	for _, a := range accs {
		cache.add(a, false)
	}
	// Add some of them twice to check that they don't get reinserted.
	cache.add(accs[0], false)
	cache.add(accs[2], true)

	// Check that the account list is sorted by filename.
	wantAccounts := make([]accountByTag, len(accs))

	for _, a := range accs {
		wantAccounts = append(wantAccounts, accountByTag{a, false})
	}
	sort.Sort(accountsByTag(wantAccounts))
	list := cache.accounts()
	if !reflect.DeepEqual(list, wantAccounts) {
		t.Fatalf("got accounts: %s\nwant %s", spew.Sdump(accs), spew.Sdump(wantAccounts))
	}
	for _, a := range accs {
		if !cache.hasAddress(a.Address) {
			t.Errorf("expected hasAccount(%x) to return true", a.Address)
		}
	}
	if cache.hasAddress(common.Base58ToAddress("2DnJDxmGkeFvvxxMP6ZUwJNQx5uGEtRfY6ahpQBkzQruouvL7sCwbgTHTFN3hernDpM2bLbUQdZds3PyUBHqaAzA")) {
		t.Errorf("expected hasAccount(%x) to return false", common.Base58ToAddress("2DnJDxmGkeFvvxxMP6ZUwJNQx5uGEtRfY6ahpQBkzQruouvL7sCwbgTHTFN3hernDpM2bLbUQdZds3PyUBHqaAzA"))
	}

	// Delete a few keys from the cache.
	for i := 0; i < len(accs); i += 2 {
		cache.delete(wantAccounts[i].accountByURL)
	}
	cache.delete(accounts.Account{Address: common.Base58ToAddress("TKtF23g7WfnTFkSbskERmaMiMxUCCgPfG4rMRGEmrRCGrUFnHULC4XA6WEF8HxGF3JbQc4f8wj8Hg9T7yv91KHs"), URL: accounts.URL{Scheme: KeyStoreScheme, Path: "something"}})

	// Check content again after deletion.
	wantAccountsAfterDelete := []accounts.Account{
		wantAccounts[1].accountByURL,
		wantAccounts[3].accountByURL,
		wantAccounts[5].accountByURL,
	}
	list = cache.accounts()
	if !reflect.DeepEqual(list, wantAccountsAfterDelete) {
		t.Fatalf("got accounts after delete: %s\nwant %s", spew.Sdump(list), spew.Sdump(wantAccountsAfterDelete))
	}
	for _, a := range wantAccountsAfterDelete {
		if !cache.hasAddress(a.Address) {
			t.Errorf("expected hasAccount(%x) to return true", a.Address)
		}
	}
	if cache.hasAddress(wantAccounts[0].accountByURL.Address) {
		t.Errorf("expected hasAccount(%x) to return false", wantAccounts[0].accountByURL.Address)
	}
}

func TestCacheFind(t *testing.T) {
	dir := filepath.Join("testdata", "dir")
	cache, _ := newAccountCache(dir)
	cache.watcher.running = true // prevent unexpected reloads

	accs := []accounts.Account{
		{
			Address: common.Base58ToAddress("2rkR2mi1WRhKdwsTkDXH1ij7jJzmJ13RMKdE5S33LAVytYrdT9EYQHUZxeoEUyEbvTzPn3E8qKYyakANyE8BLjeN"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "a.key")},
		},
		{
			Address: common.Base58ToAddress("5mLeU224vu2xzc76zVzYeFoJFn6QPb1fVnzrF9xJtV5EcqkGusFkATrmArjuiqi6mpGEdtAyRnhSo5aNmYSbJY8i"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "b.key")},
		},
		{
			Address: common.Base58ToAddress("2hxrVndUSTpuJwNspmrhDsUWCtsHStvdZVL5kkkzTuT17cvBFZhrN6txLxvsowGvbXLmTHAKYYRTPQUopoJq1hpt"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "c.key")},
		},
		{
			Address: common.Base58ToAddress("48M77Ro44VZ2a7wP5dMCrPfGJk2Fk31JiMK67R1DPicGAmW1dHUghRRrJ1GxXigFFG1dPq5vJkp7FJCz3mzHY8Dj"),
			URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "c2.key")},
		},
	}
	for _, a := range accs {
		cache.add(a, false)
	}

	nomatchAccount := accounts.Account{
		Address: common.Base58ToAddress("5hZAzZ1QuYuigswgipHTPEz2QuH4E1eEqWWGi9RznkyCq6v9uGS5DyQKyZTVWheimoqXYH4KgbrhECSDYsCauBmR"),
		URL:     accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Join(dir, "something")},
	}
	tests := []struct {
		Query      accounts.Account
		WantResult accounts.Account
		WantError  error
	}{
		// by address
		{Query: accounts.Account{Address: accs[0].Address}, WantResult: accs[0]},
		// by file
		{Query: accounts.Account{URL: accs[0].URL}, WantResult: accs[0]},
		// by basename
		{Query: accounts.Account{URL: accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Base(accs[0].URL.Path)}}, WantResult: accs[0]},
		// by file and address
		{Query: accs[0], WantResult: accs[0]},
		// ambiguous address, tie resolved by file
		{Query: accs[2], WantResult: accs[2]},
		// ambiguous address error
		{
			Query: accounts.Account{Address: accs[2].Address},
			WantError: &AmbiguousAddrError{
				Addr:    accs[2].Address,
				Matches: []accounts.Account{accs[2], accs[3]},
			},
		},
		// no match error
		{Query: nomatchAccount, WantError: ErrNoMatch},
		{Query: accounts.Account{URL: nomatchAccount.URL}, WantError: ErrNoMatch},
		{Query: accounts.Account{URL: accounts.URL{Scheme: KeyStoreScheme, Path: filepath.Base(nomatchAccount.URL.Path)}}, WantError: ErrNoMatch},
		{Query: accounts.Account{Address: nomatchAccount.Address}, WantError: ErrNoMatch},
	}
	for i, test := range tests {
		a, err := cache.find(test.Query)
		if !reflect.DeepEqual(err, test.WantError) {
			t.Errorf("test %d: error mismatch for query %v\ngot %q\nwant %q", i, test.Query, err, test.WantError)
			continue
		}
		if a != test.WantResult {
			t.Errorf("test %d: result mismatch for query %v\ngot %v\nwant %v", i, test.Query, a, test.WantResult)
			continue
		}
	}
}

func waitForAccounts(wantAccounts []accounts.Account, ks *KeyStore) error {
	var list []accounts.Account
	for d := 200 * time.Millisecond; d < 8*time.Second; d *= 2 {
		list = ks.Accounts()
		if reflect.DeepEqual(list, wantAccounts) {
			// ks should have also received change notifications
			select {
			case <-ks.changes:
			default:
				return fmt.Errorf("wasn't notified of new accounts")
			}
			return nil
		}
		time.Sleep(d)
	}
	return fmt.Errorf("\ngot  %v\nwant %v", list, wantAccounts)
}

// TestUpdatedKeyfileContents tests that updating the contents of a keystore file
// is noticed by the watcher, and the account cache is updated accordingly
func TestUpdatedKeyfileContents(t *testing.T) {
	t.Parallel()

	// Create a temporary kesytore to test with
	rand.Seed(time.Now().UnixNano())
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("sero-keystore-watch-test-%d-%d", os.Getpid(), rand.Int()))
	ks := NewKeyStore(dir, LightScryptN, LightScryptP)

	list := ks.Accounts()
	if len(list) > 0 {
		t.Error("initial account list not empty:", list)
	}
	time.Sleep(100 * time.Millisecond)

	// Create the directory and copy a key file into it.
	os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "aaa")

	// Place one of our testfiles in there
	if err := cp.CopyFile(file, cachetestAccounts[0].URL.Path); err != nil {
		t.Fatal(err)
	}

	// ks should see the account.
	wantAccounts := []accounts.Account{cachetestAccounts[0]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	if err := waitForAccounts(wantAccounts, ks); err != nil {
		t.Error(err)
		return
	}

	// needed so that modTime of `file` is different to its current value after forceCopyFile
	time.Sleep(1000 * time.Millisecond)

	// Now replace file contents
	if err := forceCopyFile(file, cachetestAccounts[1].URL.Path); err != nil {
		t.Fatal(err)
		return
	}
	wantAccounts = []accounts.Account{cachetestAccounts[1]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	if err := waitForAccounts(wantAccounts, ks); err != nil {
		t.Errorf("First replacement failed")
		t.Error(err)
		return
	}

	// needed so that modTime of `file` is different to its current value after forceCopyFile
	time.Sleep(1000 * time.Millisecond)

	// Now replace file contents again
	if err := forceCopyFile(file, cachetestAccounts[2].URL.Path); err != nil {
		t.Fatal(err)
		return
	}
	wantAccounts = []accounts.Account{cachetestAccounts[2]}
	wantAccounts[0].URL = accounts.URL{Scheme: KeyStoreScheme, Path: file}
	if err := waitForAccounts(wantAccounts, ks); err != nil {
		t.Errorf("Second replacement failed")
		t.Error(err)
		return
	}

	// needed so that modTime of `file` is different to its current value after ioutil.WriteFile
	time.Sleep(1000 * time.Millisecond)

	// Now replace file contents with crap
	if err := ioutil.WriteFile(file, []byte("foo"), 0644); err != nil {
		t.Fatal(err)
		return
	}
	if err := waitForAccounts([]accounts.Account{}, ks); err != nil {
		t.Errorf("Emptying account file failed")
		t.Error(err)
		return
	}
}

// forceCopyFile is like cp.CopyFile, but doesn't complain if the destination exists.
func forceCopyFile(dst, src string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
