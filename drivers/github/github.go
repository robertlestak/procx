package github

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v35/github"
	"github.com/google/uuid"
	"github.com/robertlestak/procx/pkg/flags"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type GitHubOp string

var (
	GitHubOpRM = GitHubOp("rm")
	GitHubOpMV = GitHubOp("mv")
)

type GitHub struct {
	Client          *github.Client
	Repo            string
	Owner           string
	Token           string
	File            string
	FilePrefix      *string
	FileRegex       *string
	Ref             *string
	OpenPR          bool
	BaseBranch      *string
	Branch          *string
	CommitName      *string
	CommitEmail     *string
	CommitMessage   *string
	PRTitle         *string
	PRBody          *string
	ClearOp         *GitHubOp
	ClearOpLocation *string
	FailOp          *GitHubOp
	FailOpLocation  *string
	data            string
}

func (d *GitHub) LoadEnv(prefix string) error {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "LoadEnv",
	})
	l.Debug("Loading environment")
	if os.Getenv(prefix+"GITHUB_REPO") != "" {
		d.Repo = os.Getenv(prefix + "GITHUB_REPO")
	}
	if os.Getenv(prefix+"GITHUB_OWNER") != "" {
		d.Owner = os.Getenv(prefix + "GITHUB_OWNER")
	}
	if os.Getenv(prefix+"GITHUB_TOKEN") != "" {
		d.Token = os.Getenv(prefix + "GITHUB_TOKEN")
	}
	if os.Getenv(prefix+"GITHUB_FILE") != "" {
		d.File = os.Getenv(prefix + "GITHUB_FILE")
	}
	if os.Getenv(prefix+"GITHUB_FILE_PREFIX") != "" {
		v := os.Getenv(prefix + "GITHUB_FILE_PREFIX")
		d.FilePrefix = &v
	}
	if os.Getenv(prefix+"GITHUB_FILE_REGEX") != "" {
		v := os.Getenv(prefix + "GITHUB_FILE_REGEX")
		d.FileRegex = &v
	}
	if os.Getenv(prefix+"GITHUB_REF") != "" {
		v := os.Getenv(prefix + "GITHUB_REF")
		d.Ref = &v
	}
	if os.Getenv(prefix+"GITHUB_OPEN_PR") != "" {
		v := os.Getenv(prefix + "GITHUB_OPEN_PR")
		d.OpenPR = v == "true"
	}
	if os.Getenv(prefix+"GITHUB_BASE_BRANCH") != "" {
		v := os.Getenv(prefix + "GITHUB_BASE_BRANCH")
		d.BaseBranch = &v
	}
	if os.Getenv(prefix+"GITHUB_BRANCH") != "" {
		v := os.Getenv(prefix + "GITHUB_BRANCH")
		d.Branch = &v
	}
	if os.Getenv(prefix+"GITHUB_COMMIT_NAME") != "" {
		v := os.Getenv(prefix + "GITHUB_COMMIT_NAME")
		d.CommitName = &v
	}
	if os.Getenv(prefix+"GITHUB_COMMIT_EMAIL") != "" {
		v := os.Getenv(prefix + "GITHUB_COMMIT_EMAIL")
		d.CommitEmail = &v
	}
	if os.Getenv(prefix+"GITHUB_COMMIT_MESSAGE") != "" {
		v := os.Getenv(prefix + "GITHUB_COMMIT_MESSAGE")
		d.CommitMessage = &v
	}
	if os.Getenv(prefix+"GITHUB_PR_TITLE") != "" {
		v := os.Getenv(prefix + "GITHUB_PR_TITLE")
		d.PRTitle = &v
	}
	if os.Getenv(prefix+"GITHUB_PR_BODY") != "" {
		v := os.Getenv(prefix + "GITHUB_PR_BODY")
		d.PRBody = &v
	}
	if os.Getenv(prefix+"GITHUB_CLEAR_OP") != "" {
		v := GitHubOp(os.Getenv(prefix + "GITHUB_CLEAR_OP"))
		d.ClearOp = &v
	}
	if os.Getenv(prefix+"GITHUB_CLEAR_OP_LOCATION") != "" {
		v := os.Getenv(prefix + "GITHUB_CLEAR_OP_LOCATION")
		d.ClearOpLocation = &v
	}
	if os.Getenv(prefix+"GITHUB_FAIL_OP") != "" {
		v := GitHubOp(os.Getenv(prefix + "GITHUB_FAIL_OP"))
		d.FailOp = &v
	}
	if os.Getenv(prefix+"GITHUB_FAIL_OP_LOCATION") != "" {
		v := os.Getenv(prefix + "GITHUB_FAIL_OP_LOCATION")
		d.FailOpLocation = &v
	}
	return nil
}

func (d *GitHub) LoadFlags() error {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "LoadFlags",
	})
	l.Debug("Loading flags")
	d.Repo = *flags.GitHubRepo
	d.Owner = *flags.GitHubOwner
	d.Token = *flags.GitHubToken
	d.File = *flags.GitHubFile
	d.FilePrefix = flags.GitHubFilePrefix
	d.FileRegex = flags.GitHubFileRegex
	d.Ref = flags.GitHubRef
	d.OpenPR = *flags.GitHubOpenPR
	d.BaseBranch = flags.GitHubBaseBranch
	d.Branch = flags.GitHubBranch
	d.CommitName = flags.GitHubCommitName
	d.CommitEmail = flags.GitHubCommitEmail
	d.CommitMessage = flags.GitHubCommitMessage
	d.PRTitle = flags.GitHubPRTitle
	d.PRBody = flags.GitHubPRBody
	o := GitHubOp(*flags.GitHubClearOp)
	d.ClearOp = &o
	d.ClearOpLocation = flags.GitHubClearOpLocation
	fo := GitHubOp(*flags.GitHubFailOp)
	d.FailOp = &fo
	d.FailOpLocation = flags.GitHubFailOpLocation
	return nil
}

func (d *GitHub) Init() error {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "Init",
	})
	l.Debug("Initializing github driver")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: d.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	d.Client = github.NewClient(tc)
	return nil
}

func (d *GitHub) GetWork() (io.Reader, error) {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "GetWork",
	})
	l.Debug("Getting work from github")
	ctx := context.Background()
	opts := &github.RepositoryContentGetOptions{}
	if d.Ref != nil && *d.Ref != "" {
		opts.Ref = *d.Ref
	}
	if d.File == "" && *d.FilePrefix != "" {
		folder := path.Dir(*d.FilePrefix)
		file := path.Base(*d.FilePrefix)
		l.WithFields(log.Fields{
			"folder": folder,
			"file":   file,
		}).Debug("Getting work from github")
		fn, err := d.findFile(folder, file, "")
		if err != nil {
			return nil, err
		}
		if fn != "" {
			d.File = fn
		} else {
			l.Debugf("No file found in %s with prefix %s", folder, file)
			return nil, nil
		}
	} else if d.File == "" && *d.FileRegex != "" {
		folder := path.Dir(*d.FileRegex)
		l.WithFields(log.Fields{
			"folder": folder,
			"file":   *d.FileRegex,
		}).Debug("Getting work from github")
		fn, err := d.findFile(".", "", *d.FileRegex)
		if err != nil {
			return nil, err
		}
		if fn != "" {
			d.File = fn
		} else {
			l.Debugf("No file found matching %s", *d.FileRegex)
			return nil, nil
		}
	}
	if d.File == "" {
		return nil, nil
	}
	contents, _, _, err := d.Client.Repositories.GetContents(ctx, d.Owner, d.Repo, d.File, opts)
	if err != nil {
		l.Debugf("GetContents error=%v", err)
		return nil, err
	}
	if contents == nil {
		l.Debugf("No contents found for %s", d.File)
		return nil, nil
	}
	c, err := contents.GetContent()
	if err != nil {
		l.Debugf("GetContent error=%v", err)
		return nil, err
	}
	d.data = c
	return strings.NewReader(c), nil
}

func (d *GitHub) ClearWork() error {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "ClearWork",
		"op":  *d.ClearOp,
	})
	l.Debug("Clearing work from github")
	if d.ClearOp == nil || *d.ClearOp == "" {
		return nil
	}
	fn := strings.ReplaceAll(*d.ClearOpLocation, "{{key}}", path.Base(d.File))
	if err := d.NewCommit(context.Background(), *d.ClearOp, fn); err != nil {
		l.Debugf("NewCommit error=%v", err)
		return err
	}
	return nil
}

func (d *GitHub) HandleFailure() error {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "ClearWork",
	})
	l.Debug("Clearing work from github")
	if d.ClearOp == nil || *d.ClearOp == "" {
		return nil
	}
	fn := strings.ReplaceAll(*d.FailOpLocation, "{{key}}", path.Base(d.File))
	if err := d.NewCommit(context.Background(), *d.FailOp, fn); err != nil {
		l.Debugf("NewCommit error=%v", err)
		return err
	}
	return nil
}

func (d *GitHub) Cleanup() error {
	return nil
}

func (d *GitHub) findFile(dir string, prefix string, regex string) (string, error) {
	l := log.WithFields(log.Fields{
		"pkg": "github",
		"fn":  "findFile",
	})
	l.Debug("Finding file")
	ctx := context.Background()
	opts := &github.RepositoryContentGetOptions{}
	if d.Ref != nil && *d.Ref != "" {
		opts.Ref = *d.Ref
	}
	var checking bool = true
	var fn string
	for d.File == "" && checking {
		_, dirContents, _, err := d.Client.Repositories.GetContents(ctx, d.Owner, d.Repo, dir, opts)
		if err != nil {
			l.Debugf("GetContents error=%v", err)
			return "", err
		}
		for i, c := range dirContents {
			l.Debugf("Found file %s %s", c.GetName(), *c.Type)
			var matcher bool
			fdir := path.Join(dir, c.GetName())
			if *d.FilePrefix != "" {
				l.Debugf("Checking prefix %s", *d.FilePrefix)
				cdir := path.Join(dir, *d.FilePrefix)
				prefixMatch := (strings.HasPrefix(fdir, cdir) || strings.HasPrefix(fdir, *d.FilePrefix))
				matcher = prefixMatch
				l.Debugf("Prefix match %v", prefixMatch)
			} else if regex != "" {
				l.Debugf("Checking regex %s", regex)
				cm, err := regexp.MatchString(regex, fdir)
				if err != nil {
					l.Debugf("Regex match error=%v", err)
					return "", err
				}
				fm, err := regexp.MatchString(regex, c.GetName())
				if err != nil {
					l.Debugf("Regex match error=%v", err)
					return "", err
				}
				matcher = cm || fm
				l.Debugf("Regex match %v", matcher)
			} else {
				return "", errors.New("no file prefix or regex specified")
			}
			if *c.Type == "file" && matcher {
				fn = fdir
				l.Debugf("Found file %s", fn)
				break
			} else if *c.Type == "dir" {
				l.Debugf("Found dir %s", fdir)
				fn, err := d.findFile(fdir, prefix, regex)
				if err != nil {
					return "", err
				}
				if fn != "" {
					d.File = fn
					break
				}
			}
			if i == len(dirContents)-1 {
				checking = false
			}
		}
		if fn != "" {
			break
		}
	}
	if fn != "" {
		d.File = fn
	}
	return d.File, nil
}

func (d *GitHub) createBranch(ctx context.Context) (*github.Reference, error) {
	l := log.WithFields(log.Fields{
		"action": "createBranch",
	})
	l.Debugf("createBranch")
	var baseRef *github.Reference
	var err error
	if d.BaseBranch == nil || *d.BaseBranch == "" {
		return nil, errors.New("no base branch specified")
	}
	if d.Branch == nil || *d.Branch == "" {
		return nil, errors.New("no branch specified")
	}
	var cref string
	if d.Ref != nil && *d.Ref != "" {
		cref = *d.Ref
	} else if d.BaseBranch != nil && *d.BaseBranch != "" {
		cref = "refs/heads/" + *d.BaseBranch
	}
	if baseRef, _, err = d.Client.Git.GetRef(ctx, d.Owner, d.Repo, cref); err != nil {
		l.Debugf("GetRef error=%v", err)
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + *d.Branch), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	var ref *github.Reference
	ref, _, err = d.Client.Git.CreateRef(ctx, d.Owner, d.Repo, newRef)
	if err != nil {
		l.Debugf("CreateRef error=%v", err)
		return nil, err
	}
	return ref, err
}

func (d *GitHub) createTree(ctx context.Context, ref *github.Reference, op GitHubOp, nl string) (*github.Tree, error) {
	l := log.WithFields(log.Fields{
		"action": "createTree",
		"op":     op,
		"ref":    ref,
	})
	l.Debugf("createTree")
	entries := []*github.TreeEntry{}
	if op == GitHubOpRM {
		te := &github.TreeEntry{
			Path: github.String(d.File),
			Type: github.String("blob"),
			Mode: github.String("100644"),
		}
		te.SHA = nil
		entries = append(entries, te)
	} else if op == GitHubOpMV {
		te := &github.TreeEntry{
			Path: github.String(d.File),
			Type: github.String("blob"),
			Mode: github.String("100644"),
		}
		te.SHA = nil
		entries = append(entries, te)
		te = &github.TreeEntry{
			Path: github.String(nl),
			Type: github.String("blob"),
			Mode: github.String("100644"),
		}
		te.SHA = nil
		te.Content = github.String(d.data)
		entries = append(entries, te)
	}
	var tree *github.Tree
	var err error
	l.Debugf("createTree: %s %+v", *ref.Object.SHA, entries)
	tree, _, err = d.Client.Git.CreateTree(ctx, d.Owner, d.Repo, *ref.Object.SHA, entries)
	if err != nil {
		l.Debugf("CreateTree error=%v", err)
		return nil, err
	}
	return tree, err
}

func (d *GitHub) pushCommit(ctx context.Context, ref *github.Reference, tree *github.Tree) (err error) {
	l := log.WithFields(log.Fields{
		"action": "pushCommit",
	})
	l.Debugf("pushCommit")
	parent, _, err := d.Client.Repositories.GetCommit(ctx, d.Owner, d.Repo, *ref.Object.SHA)
	if err != nil {
		l.Debugf("GetCommit error=%v", err)
		return err
	}
	if d.CommitEmail == nil || *d.CommitEmail == "" {
		return errors.New("no commit email specified")
	}
	if d.CommitName == nil || *d.CommitName == "" {
		return errors.New("no commit name specified")
	}
	if d.CommitMessage == nil || *d.CommitMessage == "" {
		return errors.New("no commit message specified")
	}
	parent.Commit.SHA = parent.SHA
	date := time.Now()
	author := &github.CommitAuthor{
		Date:  &date,
		Name:  github.String(*d.CommitName),
		Email: github.String(*d.CommitEmail),
	}
	commit := &github.Commit{
		Author:  author,
		Message: github.String(*d.CommitMessage),
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}
	newCommit, _, err := d.Client.Git.CreateCommit(ctx, d.Owner, d.Repo, commit)
	if err != nil {
		l.Debugf("CreateCommit error=%v", err)
		return err
	}
	ref.Object.SHA = newCommit.SHA
	_, _, err = d.Client.Git.UpdateRef(ctx, d.Owner, d.Repo, ref, false)
	if err != nil {
		l.Debugf("UpdateRef error=%v", err)
		return err
	}
	return err
}

func (d *GitHub) createPR(ctx context.Context) (string, error) {
	l := log.WithFields(log.Fields{
		"action": "createPR",
	})
	l.Debugf("createPR")
	if d.PRTitle == nil || *d.PRTitle == "" {
		d.PRTitle = d.CommitMessage
	}
	newPR := &github.NewPullRequest{
		Title:               github.String(*d.PRTitle),
		Head:                github.String(d.Owner + ":" + *d.Branch),
		Base:                github.String(*d.BaseBranch),
		Body:                github.String(*d.PRBody),
		MaintainerCanModify: github.Bool(true),
	}
	pr, _, err := d.Client.PullRequests.Create(ctx, d.Owner, d.Repo, newPR)
	if err != nil {
		l.Debugf("Create error=%v", err)
		return "", err
	}
	return pr.GetHTMLURL(), nil
}

func (d *GitHub) NewCommit(ctx context.Context, op GitHubOp, nl string) error {
	l := log.WithFields(log.Fields{
		"action": "NewCommit",
	})
	l.Debugf("New")
	var baseRef *github.Reference
	var err error
	var cref string
	if d.Ref != nil && *d.Ref != "" {
		cref = *d.Ref
	} else if d.BaseBranch != nil && *d.BaseBranch != "" {
		cref = "refs/heads/" + *d.BaseBranch
	}
	if baseRef, _, err = d.Client.Git.GetRef(ctx, d.Owner, d.Repo, cref); err != nil {
		l.Debugf("GetRef error=%v", err)
		return err
	}
	if d.OpenPR && (d.Branch == nil || *d.Branch == "") {
		b := uuid.New().String()
		d.Branch = &b
	}
	if d.Branch != nil && *d.Branch != "" {
		ref, berr := d.createBranch(ctx)
		if berr != nil {
			l.Debugf("createBranch error=%v", berr)
			return berr
		}
		baseRef = ref
	}
	tree, terr := d.createTree(ctx, baseRef, op, nl)
	if terr != nil {
		l.Debugf("createTree error=%v", terr)
		return terr
	}
	perr := d.pushCommit(ctx, baseRef, tree)
	if perr != nil {
		l.Debugf("pushCommit error=%v", perr)
		return perr
	}
	if d.OpenPR {
		purl, err := d.createPR(ctx)
		if err != nil {
			l.Debugf("createPR error=%v", err)
			return err
		}
		l.Debugf("PR created at %s", purl)
	}
	return nil
}
