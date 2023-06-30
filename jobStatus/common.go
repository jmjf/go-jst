package jobStatus

import (
	"common"
	"errors"
)

// Need to figure out what to call this.
// Repo interfaces are shared, so not part of a single repo.

type JobStatusRepo interface {
	// if running testRepo, change add() to Add() here and in the repos.
	add(jobStatus JobStatus) error
	GetByJobId(id JobIdType) ([]JobStatus, error)
	GetByJobIdBusinessDate(id JobIdType, businessDate common.Date) ([]JobStatus, error)
}

// I'm doing status these types the easy way for now.
// If I want to have database table of job statuses, I could build a map[string]int and load it.
// That's overkill for now, so keeping it simple.
type JobStatusCodeType string
type JobIdType string

const (
	JobStatus_INVALID JobStatusCodeType = "INVALID"
	JobStatus_START   JobStatusCodeType = "START"
	JobStatus_SUCCEED JobStatusCodeType = "SUCCEED"
	JobStatus_FAIL    JobStatusCodeType = "FAIL"
)

// Use validJobStatusCodes to ensure a value is a valid job status; update if job status consts change.
// INVALID is not a valid job status code, but it's defined for easy, safe checks for invalid.
var validJobStatusCodes = []JobStatusCodeType{JobStatus_START, JobStatus_SUCCEED, JobStatus_FAIL}

// Primitive errors to use with DomainError and AppError
var (
	errDomainProps    = errors.New("props error")
	codeDomainProps   = "PropsError"
	errAppUnexpected  = errors.New("unexpected error")
	codeAppUnexpected = "UnexpectedError"
	errRepoScanError  = errors.New("scan error")
	codeRepoScanError = "ScanError"
	// TODO: examine database error and classify it
	// should retry? etc.
	errRepoOtherError  = errors.New("other error")
	codeRepoOtherError = "RepoOtherError"
)
