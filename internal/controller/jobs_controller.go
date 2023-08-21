package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Running  = "Running"
	Complete = "Complete"
	Failed   = "Failed"
)

type JobsController interface {
	Completed() bool
	StartJob() error
	GetCurJob() *batchv1.Job
	CurJobStatus() (string, error)
	NextJob()
}

type jobsController struct {
	client.Client
	jobs []*batchv1.Job

	succeeds int
}

func (j *jobsController) Completed() bool {
	return j.succeeds == len(j.jobs)
}

func (j *jobsController) StartJob() error {
	if j.succeeds >= len(j.jobs) {
		return nil
	}

	job := j.jobs[j.succeeds]
	if err := j.Create(context.Background(), job); err != nil {
		// if the job already exists, ignore the error
		return client.IgnoreAlreadyExists(err)
	}

	return nil
}

func (j *jobsController) CurJobStatus() (string, error) {
	if j.succeeds >= len(j.jobs) {
		return "", nil
	}

	job := j.jobs[j.succeeds]
	if err := j.Get(context.Background(), client.ObjectKey{
		Namespace: job.Namespace,
		Name:      job.Name,
	}, job); err != nil {
		return "", err
	}

	switch {
	case job.Status.Active > 0:
		return Running, nil
	case job.Status.Succeeded > 0:
		return Complete, nil
	case job.Status.Failed > 0:
		return Failed, nil
	default:
		return "", nil
	}

}

func (j *jobsController) NextJob() {
	if j.succeeds < len(j.jobs) {
		j.succeeds++
	}
}

func (j *jobsController) GetCurJob() *batchv1.Job {
	if j.succeeds >= len(j.jobs) {
		return nil
	}

	return j.jobs[j.succeeds]
}

func NewJobsController(cli client.Client, jobs []*batchv1.Job) JobsController {
	return &jobsController{
		Client: cli,
		jobs:   jobs,
	}
}
