package zconfig

var snapshot_flag uint64

func Init_Snapshot(num uint64) {
	snapshot_flag = num
}

func IsSnapshotMode() bool {
	if snapshot_flag > 0 {
		return true
	}
	return false
}

func NeedSnapshot(num uint64) bool {
	if num == snapshot_flag {
		return true
	}
	return false
}
