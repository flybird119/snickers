package core

import (
	"path"
	"strconv"

	"code.cloudfoundry.org/lager"

	"github.com/cavaliercoder/grab"
	"github.com/snickers/snickers/db"
	"github.com/snickers/snickers/types"
)

// HTTPDownload function downloads sources using
// http protocol.
func HTTPDownload(logger lager.Logger, dbInstance db.Storage, jobID string) error {
	log := logger.Session("http-download")
	log.Info("start", lager.Data{"job": jobID})
	defer log.Info("finished")

	job, err := dbInstance.RetrieveJob(jobID)
	if err != nil {
		log.Error("retrieving-job", err)
		return err
	}

	job.LocalSource = GetLocalSourcePath(job.ID) + path.Base(job.Source)
	job.LocalDestination = GetLocalDestination(dbInstance, jobID)
	job.Destination = GetOutputFilename(dbInstance, jobID)
	job.Status = types.JobDownloading
	job.Details = "0%"
	dbInstance.UpdateJob(job.ID, job)

	respch, _ := grab.GetAsync(GetLocalSourcePath(job.ID), job.Source)

	resp := <-respch
	for !resp.IsComplete() {
		job, _ = dbInstance.RetrieveJob(jobID)
		percentage := strconv.FormatInt(int64(resp.BytesTransferred()*100/resp.Size), 10)
		if job.Details != percentage {
			job.Details = percentage + "%"
			dbInstance.UpdateJob(job.ID, job)
		}
	}

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}
