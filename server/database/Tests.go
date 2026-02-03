package database

import (
	"context"
	"strings"
)

func GetTestDetails(config string, testType string, testCategory string) (Test, bool) {
	ctx := context.Background()
	testCategory = strings.TrimSpace(testCategory)
	var test Test
	var err error
	if strings.EqualFold(testCategory, "") {
		var arg getTestWithoutCategoryParams
		arg.ConfigName = config
		arg.TestType = testType
		test, err = dbObject.getTestWithoutCategory(ctx, arg)
	} else {
		var arg getTestWithCategoryParams
		arg.ConfigName = config
		arg.TestType = testType
		arg.TestCategory = testCategory
		test, err = dbObject.getTestWithCategory(ctx, arg)
	}
	if err != nil {
		return Test{}, false
	}
	return test, true
}

func GetTestsForConfig(configName string) ([]string, bool) {
	ctx := context.Background()
	configName = strings.TrimSpace(configName)

	tsts, err := dbObject.getAllTestForConfig(ctx, configName)
	if err != nil {
		return nil, false
	}
	var tbr = make([]string, 0)
	for _, t := range tsts {
		test := t.TestType
		if !strings.EqualFold(t.TestCategory, "") {
			test = test + ";" + t.TestCategory
		}
		tbr = append(tbr, test)
	}
	return tbr, true
}

func GetConfigsForUplink() ([]string, bool) {
	ctx := context.Background()

	tsts, err := dbObject.getConfigsForUplink(ctx)
	if err != nil {
		return nil, false
	}
	return tsts, true
}
