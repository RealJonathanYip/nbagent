package reporter

type realReporter interface {
	Count(szKey string, nCount ...int64)
	Result(szKey, szResult string, nDurationAndCount ...int64)
	Duration(szKey string, nDuration uint64, Count ...int64)
}

var (
	_aryReporters []*realReporter
)

/*
	上报次数
	szKey:上报器的名字，注意唯一性
	nCount:次数默认为1
	示例：report.Count("test_report", 10) 或 report.Count("test_report")
*/
func Count(szKey string, nCount ...int64) {
	for _, ptrReporter := range _aryReporters {
		(*ptrReporter).Count(szKey, nCount...)
	}
}

/*
	上报结果
	szKey:上报器的名字，注意唯一性
	nDurationAndCount:时延（如果有可填，没有填空）以及次数（没有的话默认为1）
	示例：report.Result("test_report", "SUCCESS", 123, 2) 或 report.Result("test_report", "SUCCESS")
*/
func Result(szKey, szResult string, nDurationAndCount ...int64) {
	for _, ptrReporter := range _aryReporters {
		(*ptrReporter).Result(szKey, szResult, nDurationAndCount...)
	}
}
/*
	上报结果
	szKey:上报器的名字，注意唯一性
	nDuration:时延（如果有）以及次数（没有的话默认为1）
	nCount:次数默认为1
	示例：report.Duration("test_report", 10) 或 report.Duration("test_report", 10)
*/
func Duration(szKey string, nDuration uint64, Count ...int64) {
	for _, ptrReporter := range _aryReporters {
		(*ptrReporter).Duration(szKey, nDuration, Count...)
	}
}

func Register(ptrReporter *realReporter) {
	_aryReporters = append(_aryReporters, ptrReporter)
}