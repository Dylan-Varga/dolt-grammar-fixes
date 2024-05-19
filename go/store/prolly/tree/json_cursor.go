// Copyright 2024 Dolthub, Inc.
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

package tree

import (
	"context"
	"fmt"
	"io"
)

// JsonCursor wraps a cursor to an IndexedJsonDocument stored in a prolly tree. This allows seeking for a specific
// location in the stored JSON document.
type JsonCursor struct {
	cur         *cursor
	jsonScanner JsonScanner
}

// getPreviousKey computes the key of a cursor's predecessor node. This is important for scanning JSON because that
// key represents the location within the document where the current node begins, and can be used to compute the location
// of every value within the node.
func getPreviousKey(ctx context.Context, cur *cursor) ([]byte, error) {
	cur2 := cur.clone()
	err := cur2.retreat(ctx)
	if err != nil {
		return nil, err
	}
	if cur2.Valid() {
		return cur2.parent.CurrentKey(), nil
	}
	return nil, nil
}

// newJsonCursor takes the root node of a prolly tree representing a JSON document, and creates a new JsonCursor for reading
// JSON, starting at the specified location in the document.
func newJsonCursor(ctx context.Context, ns NodeStore, root Node, startKey jsonLocation) (*JsonCursor, error) {
	cur, err := newCursorAtKey(ctx, ns, root, startKey.key, jsonLocationOrdering{})
	if err != nil {
		return nil, err
	}
	previousKey, err := getPreviousKey(ctx, cur)
	if err != nil {
		return nil, err
	}
	jsonBytes := cur.currentValue()
	jsonDecoder := ScanJsonFromMiddleWithKey(jsonBytes, previousKey)

	jcur := JsonCursor{cur: cur, jsonScanner: jsonDecoder}
	err = jcur.AdvanceToLocation(ctx, startKey)
	if err != nil {
		return nil, err
	}
	return &jcur, nil
}

func (j JsonCursor) Valid() bool {
	return j.cur.Valid()
}

// NextValue reads and consumes an entire value from the JSON document, returning its bytes.
// Precondition: The scanner is currently at the start of a value.
func (j *JsonCursor) NextValue(ctx context.Context) (result []byte, err error) {
	if !j.jsonScanner.atStartOfValue() {
		return nil, fmt.Errorf("JSON cursor in unexpected state. This is likely a bug")
	}
	path := j.jsonScanner.currentPath
	jsonBuffer := j.jsonScanner.jsonBuffer
	startPos := j.jsonScanner.valueOffset

	parseChunk := func() {
		var crossedBoundary bool
		crossedBoundary, err = j.AdvanceToNextLocation(ctx)
		if err != nil {
			return
		}
		if crossedBoundary {
			result = append(result, jsonBuffer[startPos:]...)
			jsonBuffer = j.jsonScanner.jsonBuffer
			startPos = 0
		}
	}

	parseChunk()
	if err != nil {
		return
	}

	for compareJsonLocations(j.jsonScanner.currentPath, path) < 0 {
		parseChunk()
		if err != nil {
			return
		}
	}
	result = append(result, jsonBuffer[startPos:j.jsonScanner.valueOffset]...)
	return
}

func (j *JsonCursor) AdvanceToLocation(ctx context.Context, path jsonLocation) error {
	// TODO: Intelligently determine whether the location exists in the current block, or whether we need to fetch a new block.
	// Note for anyone debugging this: An infinite loop here means we're scanning the same block over and over again
	// and not finding the requested location.

	for compareJsonLocations(j.jsonScanner.currentPath, path) < 0 {
		err := j.jsonScanner.AdvanceToNextLocation()
		if err == io.EOF {
			// We reached the end of the chunk, advance the parent cursor.
			err = Seek(ctx, j.cur.parent, path.key, jsonLocationOrdering{})
			if err != nil {
				return err
			}
			j.cur.nd, err = fetchChild(ctx, j.cur.nrw, j.cur.parent.currentRef())
			if err != nil {
				return err
			}
			previousKey, err := getPreviousKey(ctx, j.cur)
			if err != nil {
				return err
			}
			j.jsonScanner = ScanJsonFromMiddleWithKey(j.cur.currentValue(), previousKey)
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (j *JsonCursor) AdvanceToNextLocation(ctx context.Context) (crossedBoundary bool, err error) {
	for {
		err = j.jsonScanner.AdvanceToNextLocation()
		if err == io.EOF {
			crossedBoundary = true
			// We hit the end of the chunk, load the next one
			err = j.cur.advance(ctx)
			if err != nil {
				return
			}
			if !j.cur.Valid() {
				// We hit the end of the tree.
				// TODO: What is the correct behavior here?
				return crossedBoundary, io.EOF
			}
			j.jsonScanner = ScanJsonFromMiddle(j.cur.currentValue(), j.jsonScanner.currentPath)
			continue
		} else if err != nil {
			return
		}
		return
	}
}

func (j *JsonCursor) GetCurrentPath() jsonLocation {
	return j.jsonScanner.currentPath
}

func (j *JsonCursor) nextCharacter() byte {
	return j.jsonScanner.jsonBuffer[j.jsonScanner.valueOffset]
}
