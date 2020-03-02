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

var testForkStartBlock uint64

func Init_TestForkStartBlock(num uint64) {
	testForkStartBlock = num
}

func GetTestForkStartBlock() uint64 {
	return testForkStartBlock
}

var testFork bool

func Init_TestFork() {
	testFork = true
}

func IsTestFork() bool {
	if testFork && testForkStartBlock == 0 {
		panic("please set testForkStartBlock")
	}
	return testFork
}

var recordBlockShareNumber bool

func Init_RecordShareNum() {
	recordBlockShareNumber = true
}

func RecordShareNum() bool {
	return recordBlockShareNumber
}
