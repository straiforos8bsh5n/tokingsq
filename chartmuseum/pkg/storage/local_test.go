package storage

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type LocalTestSuite struct {
	suite.Suite
	LocalFilesystemBackend *LocalFilesystemBackend
	BrokenTempDirectory    string
}

func (suite *LocalTestSuite) SetupSuite() {
	timestamp := time.Now().Format("20060102150405")
	suite.BrokenTempDirectory = fmt.Sprintf("../../.test/storage-local/%s-broken", timestamp)
	defer os.RemoveAll(suite.BrokenTempDirectory)
	backend := NewLocalFilesystemBackend(suite.BrokenTempDirectory)
	suite.LocalFilesystemBackend = backend
}

func (suite *LocalTestSuite) TestListObjects() {
	_, err := suite.LocalFilesystemBackend.ListObjects("")
	suite.Nil(err, "list objects does not return error if dir does not exist")
}

func (suite *LocalTestSuite) TestGetObject() {
	_, err := suite.LocalFilesystemBackend.GetObject("this-file-cannot-possibly-exist.tgz")
	suite.NotNil(err, "cannot get objects with bad path")
}

func (suite *LocalTestSuite) TestPutObjectWithNonExistentPath() {
	err := suite.LocalFilesystemBackend.PutObject("testdir/test/test.tgz", []byte("test content"))
	suite.Nil(err)
}

func TestLocalStorageTestSuite(t *testing.T) {
	suite.Run(t, new(LocalTestSuite))
}
