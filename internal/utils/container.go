package utils

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"github.com/apecloud/kubebench/api/v1alpha1"
	"github.com/apecloud/kubebench/pkg/constants"
)

// InitPGDatabaseContainer will create a database in postgresql
func InitPGDatabaseContainer(target v1alpha1.Target, database string) *corev1.Container {
	args := []string{
		"postgresql",
		"create",
		database,
		"--host", target.Host,
		"--port", strconv.Itoa(target.Port),
		"--user", target.User,
		"--password", target.Password,
	}

	return &corev1.Container{
		Name:            "init",
		Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/tools"},
		Args:            args,
	}
}

// InitMysqlDatabaseContainer will create a database in mysql
func InitMysqlDatabaseContainer(target v1alpha1.Target, database string) *corev1.Container {
	args := []string{
		"mysql",
		"create",
		database,
		"--host", target.Host,
		"--port", strconv.Itoa(target.Port),
		"--user", target.User,
		"--password", target.Password,
	}

	return &corev1.Container{
		Name:            "init",
		Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/tools"},
		Args:            args,
	}
}

func CleanMysqlDatabaseContainer(target v1alpha1.Target, database string) *corev1.Container {
	args := []string{
		"mysql",
		"drop",
		database,
		"--host", target.Host,
		"--port", strconv.Itoa(target.Port),
		"--user", target.User,
		"--password", target.Password,
	}

	return &corev1.Container{
		Name:            "clean",
		Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/tools"},
		Args:            args,
	}
}

func CleanPGDatabaseContainer(target v1alpha1.Target, database string) *corev1.Container {
	args := []string{
		"postgresql",
		"drop",
		database,
		"--host", target.Host,
		"--port", strconv.Itoa(target.Port),
		"--user", target.User,
		"--password", target.Password,
	}

	return &corev1.Container{
		Name:            "clean",
		Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/tools"},
		Args:            args,
	}
}

func CleanMongoDatabaseContainer(target v1alpha1.Target, database string) *corev1.Container {
	args := []string{
		"mongodb",
		"drop",
		database,
		"--host", target.Host,
		"--port", strconv.Itoa(target.Port),
		"--user", target.User,
		"--password", target.Password,
	}

	return &corev1.Container{
		Name:            "clean",
		Image:           constants.GetBenchmarkImage(constants.KubebenchTools),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/tools"},
		Args:            args,
	}
}
