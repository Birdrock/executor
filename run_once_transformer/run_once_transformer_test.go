package run_once_transformer_test

import (
	"github.com/cloudfoundry-incubator/executor/actionrunner/downloader"
	"github.com/cloudfoundry-incubator/executor/actionrunner/downloader/fakedownloader"
	"github.com/cloudfoundry-incubator/executor/actionrunner/logstreamer/fakelogstreamer"
	"github.com/cloudfoundry-incubator/executor/actionrunner/uploader/fakeuploader"
	. "github.com/cloudfoundry-incubator/executor/run_once_transformer"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	steno "github.com/cloudfoundry/gosteno"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vito/gordon/fake_gordon"

	"github.com/cloudfoundry-incubator/executor/action_runner"
	"github.com/cloudfoundry-incubator/executor/actionrunner/uploader"
	"github.com/cloudfoundry-incubator/executor/backend_plugin"
	"github.com/cloudfoundry-incubator/executor/linuxplugin"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/execute_action/download_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/execute_action/fetch_result_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/execute_action/run_action"
	"github.com/cloudfoundry-incubator/executor/runoncehandler/execute_action/upload_action"
)

var _ = Describe("RunOnceTransformer", func() {
	var (
		backendPlugin      backend_plugin.BackendPlugin
		downloader         downloader.Downloader
		logger             *steno.Logger
		streamer           *fakelogstreamer.FakeLogStreamer
		uploader           uploader.Uploader
		wardenClient       *fake_gordon.FakeGordon
		runOnceTransformer *RunOnceTransformer
	)

	BeforeEach(func() {
		backendPlugin = linuxplugin.New()
		downloader = &fakedownloader.FakeDownloader{}
		uploader = &fakeuploader.FakeUploader{}
		logger = &steno.Logger{}
		runOnceTransformer = NewRunOnceTransformer(
			streamer,
			downloader,
			uploader,
			backendPlugin,
			wardenClient,
			logger,
			"/fake/temp/dir",
		)
	})

	It("is correct", func() {
		runActionModel := models.RunAction{Script: "do-something"}
		downloadActionModel := models.DownloadAction{From: "/file/to/download"}
		uploadActionModel := models.UploadAction{From: "/file/to/upload"}
		fetchResultActionModel := models.FetchResultAction{File: "some-file"}

		runOnce := models.RunOnce{
			Guid: "some-guid",
			Actions: []models.ExecutorAction{
				{runActionModel},
				{downloadActionModel},
				{uploadActionModel},
				{fetchResultActionModel},
			},
			ContainerHandle: "some-container-handle",
		}

		Ω(runOnceTransformer.ActionsFor(&runOnce)).To(Equal([]action_runner.Action{
			run_action.New(
				runActionModel,
				"some-container-handle",
				streamer,
				backendPlugin,
				wardenClient,
				logger,
			),
			download_action.New(
				downloadActionModel,
				"some-container-handle",
				downloader,
				"/fake/temp/dir",
				backendPlugin,
				wardenClient,
				logger,
			),
			upload_action.New(
				uploadActionModel,
				"some-container-handle",
				uploader,
				"/fake/temp/dir",
				wardenClient,
				logger,
			),
			fetch_result_action.New(
				&runOnce,
				fetchResultActionModel,
				"some-container-handle",
				"/fake/temp/dir",
				wardenClient,
				logger,
			),
		}))
	})
})
