package gdnative

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func WithVersionMismatch(expectedMajor, expectedMinor, actualMajor, actualMinor int) logrus.Fields {
	expected := fmt.Sprintf("%d.%d", expectedMajor, expectedMinor)
	actual := fmt.Sprintf("%d.%d", actualMajor, actualMinor)

	return logrus.Fields{
		"expected": expected,
		"actual":   actual,
	}
}
