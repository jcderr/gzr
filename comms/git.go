package comms

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// GitManager is an interface to retrieve data from a git repo
type GitManager interface {
	CommitHash() (string, error)
	Remote() (string, error)
	Tags() ([]string, []string, error)
	SetPath(string)
	GetPath() string
}

// LocalGitManager implements GitManager
type LocalGitManager struct {
	path string
}

// NewImageMetadata returns a populated ImageMetadata based on a LocalGitManager
func NewImageMetadata() (ImageMetadata, error) {
	meta := ImageMetadata{}
	path, err := os.Getwd()
	if err != nil {
		return meta, err
	}
	gm := NewLocalGitManager(path)
	tags, annotations, err := gm.Tags()
	if err != nil {
		return meta, err
	}
	meta.GitTag = tags
	meta.GitAnnotation = annotations

	remote, err := gm.Remote()
	if err != nil {
		return meta, err
	}
	meta.GitOrigin = remote

	hash, err := gm.CommitHash()
	if err != nil {
		return meta, err
	}
	meta.GitCommit = hash

	meta.CreatedAt = time.Now().Format(time.RFC3339)

	return meta, nil
}

// NewLocalGitManager returns a pointer to an intialized LocalGitManager and takes a `path`
func NewLocalGitManager(path ...string) *LocalGitManager {
	var thePath string
	if path != nil {
		thePath = path[0]
	}
	return &LocalGitManager{path: thePath}
}

// CommitHash returns the commit hash of a git repo at either the set path or current
// working directory
func (gm *LocalGitManager) CommitHash() (string, error) {
	if gm.path != "" {
		oldPath, err := os.Getwd()
		if err != nil {
			return "", err
		}
		err = os.Chdir(gm.path)
		if err != nil {
			return "", err
		}
		defer os.Chdir(oldPath)
	}
	hashCmd := exec.Command("git", "rev-parse", "HEAD")

	hash, err := hashCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	stripped := strings.TrimSpace(string(hash))
	return stripped, nil
}

// Remote returns the remote of a git repo at either the set path or current
// working directory
func (gm *LocalGitManager) Remote() (string, error) {
	if gm.path != "" {
		oldPath, err := os.Getwd()
		if err != nil {
			return "", err
		}
		err = os.Chdir(gm.path)
		if err != nil {
			return "", err
		}
		defer os.Chdir(oldPath)
	}
	remoteCmd := exec.Command("git", "remote", "-v")
	remote, err := remoteCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	remotes := strings.Fields(string(remote))
	if len(remotes) == 0 {
		return "", nil
	}
	return remotes[1], nil
}

// Tags returns the tags and accompanying annotations of a git repo at either
// the set path or current working directory
func (gm *LocalGitManager) Tags() ([]string, []string, error) {
	if gm.path != "" {
		oldPath, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}
		err = os.Chdir(gm.path)
		if err != nil {
			return nil, nil, err
		}
		defer os.Chdir(oldPath)
	}
	var tags []string
	var annotations []string
	currentTags := exec.Command("git", "tag", "--format", "%(refname:strip=2)~%(contents:subject)", "-l", "-n1", "--points-at", "HEAD")

	tagInfo, err := currentTags.CombinedOutput()
	if err != nil {
		return nil, nil, err
	}
	regex, _ := regexp.Compile("\n\n")
	tagInfo_ := regex.ReplaceAllString(string(tagInfo), "\n")

	splitTagInfo := strings.Split(tagInfo_, "\n")
	for _, v := range splitTagInfo {
		if v == "" {
			continue
		}
		split := strings.Split(v, "~")
		if len(split) == 2 {
			tag := split[0]
			annotation := split[1]
			tags = append(tags, tag)
			annotations = append(annotations, annotation)
		}
		if len(split) == 1 {
			tag := split[0]
			tags = append(tags, tag)
		}
	}
	return tags, annotations, nil
}