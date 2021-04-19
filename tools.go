package main

import (
	"strconv"
)

func getInt64(data interface{}) int64 {
	switch n := data.(type) {
	case float64:
		return int64(n)
	case uint64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case string:
		num, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return -1
		}
		return num
	case nil:
		return 0
	case bool:
		if n {
			return 1
		}
		return 0
	}
	return -1
}

func getInt(data interface{}) int {
	switch n := data.(type) {
	case float64:
		return int(n)
	case uint64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		num, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			return -1
		}
		return int(num)
	case nil:
		return 0
	case bool:
		if n {
			return 1
		}
		return 0
	}
	return -1
}

func getFloat64(data interface{}) float64 {
	switch n := data.(type) {
	case string:
		num, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return -1
		}
		return num
	case float64:
		return n
	case int64:
		return float64(n)
	case uint64:
		return float64(n)
	case int:
		return float64(n)
	case nil:
		return 0
	case bool:
		if n {
			return 1
		}
		return 0
	}
	return -1
}

func getString(data interface{}) string {
	switch n := data.(type) {
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case uint64:
		return strconv.FormatUint(n, 10)
	case int:
		return strconv.Itoa(n)
	case int64:
		return strconv.FormatInt(n, 10)
	case string:
		return n
	case bool:
		if n {
			return "1"
		}
		return "0"
	case nil:
		return "0"
	}
	return "-1"
}
