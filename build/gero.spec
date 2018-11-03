# Don't try fancy stuff like debuginfo, which is useless on binary-only
# packages. Don't strip binary too
# Be sure buildpolicy set to do nothing
%define        _topdir %(pwd)
%define   _tmpdir %{_topdir}/tmp
%define        __spec_install_post %{nil}
%define          debug_package %{nil}
%define        __os_install_post %{_dbpath}/brp-compress

Summary: A very simple toy bin rpm package
Name: gero
Version: 1.0
Release: 1
License: GPL+
Group: Development/Tools
SOURCE0 : %{name}-%{version}.tar.gz
URL: http://gero.sero.com/

BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}-root
AutoReq : yes
AutoReqProv : yes
AutoProv : yes
#Requires: GLIBC_2.4 GCC_3.0 GOMP_1.0 GLIBCXX_3.4.21 GNU_HASH
#BuildRequires: glibc >= 2.4 
#BuildRequires: gcc >= 3.0 
#BuildRequires: gomp >= 1.0 
#BuildRequires: glibcxx >= 3.4.21 gnu_hash
#BuildRequires: gnu_hash
%description
%{summary}

%prep
%setup -q

%build
# Empty section.
%install
rm -rf %{buildroot}
mkdir -p  %{buildroot}

# in builddir
cp -a * %{buildroot}


%clean
#rm -rf %{buildroot}


%files
%defattr(-,root,root,-)
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.conf
%{_bindir}/../local/*
%{_libdir}/*
%changelog
* Sat Nov 03 2018  riverwind<riverwind@gameternal.com> 1.0-1
- First Build

