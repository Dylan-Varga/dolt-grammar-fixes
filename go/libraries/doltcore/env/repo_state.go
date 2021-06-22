// Copyright 2019 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dolthub/dolt/go/cmd/dolt/errhand"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdocs"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/store/hash"
)

type RepoStateReader interface {
	CWBHeadRef() ref.DoltRef
	CWBHeadSpec() *doltdb.CommitSpec
	CWBHeadHash(ctx context.Context) (hash.Hash, error)
	WorkingRoot(ctx context.Context) (*doltdb.RootValue, error)
	StagedHash() hash.Hash
	IsMergeActive() bool
	GetMergeCommit() string
	GetPreMergeWorking() string
}

type RepoStateWriter interface {
	// SetCWBHeadSpec(context.Context, *doltdb.CommitSpec) error
	UpdateStagedRoot(ctx context.Context, newRoot *doltdb.RootValue) error
	UpdateWorkingRoot(ctx context.Context, newRoot *doltdb.RootValue) error
	SetCWBHeadRef(context.Context, ref.MarshalableRef) error
	AbortMerge() error
	ClearMerge() error
	StartMerge(commitStr string) error
}

type DocsReadWriter interface {
	// GetDocsOnDisk returns the docs in the filesytem optionally filtered by docNames.
	GetDocsOnDisk(docNames ...string) (doltdocs.Docs, error)
	// WriteDocsToDisk updates the documents stored in the filesystem with the contents in docs.
	WriteDocsToDisk(docs doltdocs.Docs) error
}

type DbData struct {
	Ddb *doltdb.DoltDB
	Rsw RepoStateWriter
	Rsr RepoStateReader
	Drw DocsReadWriter
}

type BranchConfig struct {
	Merge  ref.MarshalableRef `json:"head"`
	Remote string             `json:"remote"`
}

type MergeState struct {
	Commit          string `json:"commit"`
	PreMergeWorking string `json:"working_pre_merge"`
}

type RepoState struct {
	Head     ref.MarshalableRef      `json:"head"`
	Merge    *MergeState             `json:"merge"`
	Remotes  map[string]Remote       `json:"remotes"`
	Branches map[string]BranchConfig `json:"branches"`
	// staged and working are legacy fields left over from when Dolt repos stored this info in the repo state file, not
	// in the DB directly. They're still here so that we can migrate existing repositories forward to the new storage
	// format, but they should be used only for this purpose and are no longer written.
	staged   string                  `json:"staged"`
	working  string                  `json:"working"`
}

func LoadRepoState(fs filesys.ReadWriteFS) (*RepoState, error) {
	path := getRepoStateFile()
	data, err := fs.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var repoState RepoState
	err = json.Unmarshal(data, &repoState)

	if err != nil {
		return nil, err
	}

	return &repoState, nil
}

func CloneRepoState(fs filesys.ReadWriteFS, r Remote) (*RepoState, error) {
	h := hash.Hash{}
	hashStr := h.String()
	rs := &RepoState{Head: ref.MarshalableRef{
		Ref: ref.NewBranchRef("master")},
		staged:   hashStr,
		working:  hashStr,
		Remotes:  map[string]Remote{r.Name: r},
		Branches: make(map[string]BranchConfig),
	}

	err := rs.Save(fs)

	if err != nil {
		return nil, err
	}

	return rs, nil
}

func CreateRepoState(fs filesys.ReadWriteFS, br string, rootHash hash.Hash) (*RepoState, error) {
	headRef, err := ref.Parse(br)

	if err != nil {
		return nil, err
	}

	rs := &RepoState{
		Head:     ref.MarshalableRef{Ref: headRef},
		Remotes:  make(map[string]Remote),
		Branches: make(map[string]BranchConfig),
	}

	err = rs.Save(fs)

	if err != nil {
		return nil, err
	}

	return rs, nil
}

func (rs *RepoState) Save(fs filesys.ReadWriteFS) error {
	data, err := json.MarshalIndent(rs, "", "  ")

	if err != nil {
		return err
	}

	path := getRepoStateFile()

	return fs.WriteFile(path, data)
}

func (rs *RepoState) CWBHeadRef() ref.DoltRef {
	return rs.Head.Ref
}

func (rs *RepoState) CWBHeadSpec() *doltdb.CommitSpec {
	spec, _ := doltdb.NewCommitSpec("HEAD")
	return spec
}

func (rs *RepoState) StartMerge(commit string, fs filesys.Filesys) error {
	rs.Merge = &MergeState{commit, rs.working}
	return rs.Save(fs)
}

func (rs *RepoState) AbortMerge(fs filesys.Filesys) error {
	rs.working = rs.Merge.PreMergeWorking
	return rs.ClearMerge(fs)
}

func (rs *RepoState) ClearMerge(fs filesys.Filesys) error {
	rs.Merge = nil
	return rs.Save(fs)
}

func (rs *RepoState) AddRemote(r Remote) {
	rs.Remotes[r.Name] = r
}

func (rs *RepoState) IsMergeActive() bool {
	return rs.Merge != nil
}

func (rs *RepoState) GetMergeCommit() string {
	return rs.Merge.Commit
}

// Updates the working root.
func UpdateWorkingRoot(ctx context.Context, rsw RepoStateWriter, newRoot *doltdb.RootValue) error {
	//logrus.Infof("Updating working root with value %s", newRoot.DebugString(ctx, true))

	err := rsw.UpdateWorkingRoot(ctx, newRoot)
	if err != nil {
		return ErrStateUpdate
	}

	return nil
}

// Returns the head root.
func HeadRoot(ctx context.Context, ddb *doltdb.DoltDB, rsr RepoStateReader) (*doltdb.RootValue, error) {
	commit, err := ddb.ResolveCommitRef(ctx, rsr.CWBHeadRef())

	if err != nil {
		return nil, err
	}

	return commit.GetRootValue()
}

// Returns the staged root.
func StagedRoot(ctx context.Context, ddb *doltdb.DoltDB, rsr RepoStateReader) (*doltdb.RootValue, error) {
	return ddb.ReadRootValue(ctx, rsr.StagedHash())
}

// Updates the staged root.
func UpdateStagedRoot(ctx context.Context, ddb *doltdb.DoltDB, rsw RepoStateWriter, newRoot *doltdb.RootValue) error {
	err := rsw.UpdateStagedRoot(ctx, newRoot)
	if err != nil {
		return ErrStateUpdate
	}

	return nil
}

func UpdateStagedRootWithVErr(ddb *doltdb.DoltDB, rsw RepoStateWriter, updatedRoot *doltdb.RootValue) errhand.VerboseError {
	err := UpdateStagedRoot(context.Background(), ddb, rsw, updatedRoot)

	switch err {
	case doltdb.ErrNomsIO:
		return errhand.BuildDError("fatal: failed to write value").Build()
	case ErrStateUpdate:
		return errhand.BuildDError("fatal: failed to update the staged root state").Build()
	}

	return nil
}

// TODO: this needs to be a function in the merge package, not repo state
func MergeWouldStompChanges(ctx context.Context, workingRoot *doltdb.RootValue, mergeCommit *doltdb.Commit, dbData DbData) ([]string, map[string]hash.Hash, error) {
	headRoot, err := HeadRoot(ctx, dbData.Ddb, dbData.Rsr)
	if err != nil {
		return nil, nil, err
	}

	mergeRoot, err := mergeCommit.GetRootValue()
	if err != nil {
		return nil, nil, err
	}

	headTableHashes, err := mapTableHashes(ctx, headRoot)
	if err != nil {
		return nil, nil, err
	}

	workingTableHashes, err := mapTableHashes(ctx, workingRoot)
	if err != nil {
		return nil, nil, err
	}

	mergeTableHashes, err := mapTableHashes(ctx, mergeRoot)
	if err != nil {
		return nil, nil, err
	}

	headWorkingDiffs := diffTableHashes(headTableHashes, workingTableHashes)
	mergedHeadDiffs := diffTableHashes(headTableHashes, mergeTableHashes)

	stompedTables := make([]string, 0, len(headWorkingDiffs))
	for tName, _ := range headWorkingDiffs {
		if _, ok := mergedHeadDiffs[tName]; ok {
			// even if the working changes match the merge changes, don't allow (matches git behavior).
			stompedTables = append(stompedTables, tName)
		}
	}

	return stompedTables, headWorkingDiffs, nil
}

// GetGCKeepers queries |rsr| to find a list of values that need to be temporarily saved during GC.
func GetGCKeepers(ctx context.Context, rsr RepoStateReader, ddb *doltdb.DoltDB) ([]hash.Hash, error) {
	workingRoot, err := rsr.WorkingRoot(ctx)
	if err != nil {
		return nil, err
	}

	workingHash, err := workingRoot.HashOf()
	if err != nil {
		return nil, err
	}

	keepers := []hash.Hash{
		workingHash,
		rsr.StagedHash(),
	}

	if rsr.IsMergeActive() {
		spec, err := doltdb.NewCommitSpec(rsr.GetMergeCommit())
		if err != nil {
			return nil, err
		}

		cm, err := ddb.Resolve(ctx, spec, nil)
		if err != nil {
			return nil, err
		}

		ch, err := cm.HashOf()
		if err != nil {
			return nil, err
		}

		pmw := hash.Parse(rsr.GetPreMergeWorking())
		val, err := ddb.ValueReadWriter().ReadValue(ctx, pmw)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, fmt.Errorf("MergeState.PreMergeWorking is a dangling hash")
		}

		keepers = append(keepers, ch, pmw)
	}

	return keepers, nil
}

func mapTableHashes(ctx context.Context, root *doltdb.RootValue) (map[string]hash.Hash, error) {
	names, err := root.GetTableNames(ctx)

	if err != nil {
		return nil, err
	}

	nameToHash := make(map[string]hash.Hash)
	for _, name := range names {
		h, ok, err := root.GetTableHash(ctx, name)

		if err != nil {
			return nil, err
		} else if !ok {
			panic("GetTableNames returned a table that GetTableHash says isn't there.")
		} else {
			nameToHash[name] = h
		}
	}

	return nameToHash, nil
}

func diffTableHashes(headTableHashes, otherTableHashes map[string]hash.Hash) map[string]hash.Hash {
	diffs := make(map[string]hash.Hash)
	for tName, hh := range headTableHashes {
		if h, ok := otherTableHashes[tName]; ok {
			if h != hh {
				// modification
				diffs[tName] = h
			}
		} else {
			// deletion
			diffs[tName] = hash.Hash{}
		}
	}

	for tName, h := range otherTableHashes {
		if _, ok := headTableHashes[tName]; !ok {
			// addition
			diffs[tName] = h
		}
	}

	return diffs
}
