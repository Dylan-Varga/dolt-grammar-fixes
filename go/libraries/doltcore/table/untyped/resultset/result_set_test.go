package resultset

import (
	"fmt"
	"github.com/attic-labs/noms/go/types"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/row"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/rowconv"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/ld/dolt/go/libraries/doltcore/table/untyped"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConcatSchemas(t *testing.T) {
	t.Run("Test concat all schemas", func(t *testing.T) {
		colColl := getAllSchemaColumns(t)

		expectedSch := schema.UnkeyedSchemaFromCols(colColl)
		sch, _ := ConcatSchemas(
			untyped.UnkeySchema(peopleTestSchema),
			untyped.UnkeySchema(episodesTestSchema),
			untyped.UnkeySchema(appearancesTestSchema))

		assert.Equal(t, expectedSch, sch)
	})

	t.Run("Test concat two schemas", func(t *testing.T) {
		colColl, _ := schema.NewColCollection(
			schema.NewColumn("id", 0, types.IntKind, true),
			schema.NewColumn("first", 1, types.StringKind, false),
			schema.NewColumn("last", 2, types.StringKind, false),
			schema.NewColumn("is_married", 3, types.BoolKind, false),
			schema.NewColumn("age", 4, types.IntKind, false),
			schema.NewColumn("rating", 5, types.FloatKind, false),
			schema.NewColumn("uuid", 6, types.UUIDKind, false),
			schema.NewColumn("num_episodes", 7, types.UintKind, false),
			schema.NewColumn("character_id", 8, types.IntKind, true),
			schema.NewColumn("episode_id", 9, types.IntKind, true),
			schema.NewColumn("comments", 10, types.StringKind, false),
		)

		expectedSch := schema.UnkeyedSchemaFromCols(colColl)
		sch, _ := ConcatSchemas(
			untyped.UnkeySchema(peopleTestSchema),
			untyped.UnkeySchema(appearancesTestSchema))
		assert.Equal(t, expectedSch, sch)
	})

	t.Run("Test concat one schema", func(t *testing.T) {
		colColl, _ := schema.NewColCollection(
			schema.NewColumn("id", 0, types.IntKind, true),
			schema.NewColumn("first", 1, types.StringKind, false),
			schema.NewColumn("last", 2, types.StringKind, false),
			schema.NewColumn("is_married", 3, types.BoolKind, false),
			schema.NewColumn("age", 4, types.IntKind, false),
			schema.NewColumn("rating", 5, types.FloatKind, false),
			schema.NewColumn("uuid", 6, types.UUIDKind, false),
			schema.NewColumn("num_episodes", 7, types.UintKind, false),
		)

		expectedSch := schema.UnkeyedSchemaFromCols(colColl)
		sch, _ := ConcatSchemas(untyped.UnkeySchema(peopleTestSchema))
		assert.Equal(t, expectedSch, sch)
	})
}

func TestSubsetSchema(t *testing.T) {
	t.Run("Test subset schema", func(t *testing.T) {
		colColl, _ := schema.NewColCollection(
			schema.NewColumn("id", idTag, types.IntKind, true),
			schema.NewColumn("last", lastTag, types.StringKind, false),
			schema.NewColumn("age", ageTag, types.IntKind, false),
			schema.NewColumn("uuid", uuidTag, types.UUIDKind, false),
		)

		expectedSch := schema.UnkeyedSchemaFromCols(colColl)
		sch := SubsetSchema(peopleTestSchema, "id", "last", "age", "uuid")
		assert.Equal(t, expectedSch, sch)
	})

	t.Run("Test subset unknown column", func(t *testing.T) {
		assert.Panics(t, func() {
			SubsetSchema(peopleTestSchema, "unknown")
		})
	})

	t.Run("Test subset no columns", func(t *testing.T) {
		expectedSch := schema.UnkeyedSchemaFromCols(schema.EmptyColColl)
		sch := SubsetSchema(peopleTestSchema)
		assert.Equal(t, expectedSch, sch)
	})
}

func TestNewFromSchema(t *testing.T) {
	t.Run("Test combine all schemas", func(t *testing.T) {
		colColl := getAllSchemaColumns(t)
		destSch := schema.UnkeyedSchemaFromCols(colColl)

		rss, err := NewFromDestSchema(destSch)
		assert.Nil(t, err)

		assert.Nil(t, rss.AddSchema(peopleTestSchema))
		assert.Nil(t, rss.AddSchema(episodesTestSchema))
		assert.Nil(t, rss.AddSchema(appearancesTestSchema))

		peopleToDestMapping := map[uint64]uint64{
			0: 0,
			1: 1,
			2: 2,
			3: 3,
			4: 4,
			// missing 5 tag in people schema
			6: 5,
			7: 6,
			8: 7,
		}
		episodesToDestMapping := map[uint64]uint64{
			0: 8,
			1: 9,
			2: 10,
			3: 11,
		}
		appsToDestMapping := map[uint64]uint64{
			0: 12,
			1: 13,
			2: 14,
		}

		expectedMapping := make(map[schema.Schema]*rowconv.FieldMapping)
		expectedMapping[peopleTestSchema], err = rowconv.NewFieldMapping(peopleTestSchema, destSch, peopleToDestMapping)
		assert.Nil(t, err)
		expectedMapping[episodesTestSchema], err = rowconv.NewFieldMapping(episodesTestSchema, destSch, episodesToDestMapping)
		assert.Nil(t, err)
		expectedMapping[appearancesTestSchema], err = rowconv.NewFieldMapping(appearancesTestSchema, destSch, appsToDestMapping)
		assert.Nil(t, err)

		assert.Equal(t, expectedMapping, rss.mapping)
		assert.Equal(t, destSch, rss.destSch)
	})

	t.Run("Test combine all schemas, opposite order", func(t *testing.T) {
		colColl, err := schema.NewColCollection(
			schema.NewColumn("character_id", 0, types.IntKind, true),
			schema.NewColumn("episode_id", 1, types.IntKind, true),
			schema.NewColumn("comments", 2, types.StringKind, false),
			schema.NewColumn("id", 3, types.IntKind, true),
			schema.NewColumn("name", 4, types.StringKind, false),
			schema.NewColumn("air_date", 5, types.IntKind, false),
			schema.NewColumn("rating", 6, types.FloatKind, false),
			schema.NewColumn("id", 7, types.IntKind, true),
			schema.NewColumn("first", 8, types.StringKind, false),
			schema.NewColumn("last", 9, types.StringKind, false),
			schema.NewColumn("is_married", 10, types.BoolKind, false),
			schema.NewColumn("age", 11, types.IntKind, false),
			schema.NewColumn("rating", 12, types.FloatKind, false),
			schema.NewColumn("uuid", 13, types.UUIDKind, false),
			schema.NewColumn("num_episodes", 14, types.UintKind, false),
		)
		assert.Nil(t, err)
		destSch := schema.UnkeyedSchemaFromCols(colColl)

		rss, err := NewFromDestSchema(destSch)
		assert.Nil(t, err)

		assert.Nil(t, rss.AddSchema(appearancesTestSchema))
		assert.Nil(t, rss.AddSchema(episodesTestSchema))
		assert.Nil(t, rss.AddSchema(peopleTestSchema))

		appsToDestMapping := map[uint64]uint64{
			0: 0,
			1: 1,
			2: 2,
		}
		episodesToDestMapping := map[uint64]uint64{
			0: 3,
			1: 4,
			2: 5,
			3: 6,
		}
		peopleToDestMapping := map[uint64]uint64{
			0: 7,
			1: 8,
			2: 9,
			3: 10,
			4: 11,
			// missing 5 tag in people schema
			6: 12,
			7: 13,
			8: 14,
		}

		expectedMapping := make(map[schema.Schema]*rowconv.FieldMapping)
		expectedMapping[peopleTestSchema], err = rowconv.NewFieldMapping(peopleTestSchema, destSch, peopleToDestMapping)
		assert.Nil(t, err)
		expectedMapping[episodesTestSchema], err = rowconv.NewFieldMapping(episodesTestSchema, destSch, episodesToDestMapping)
		assert.Nil(t, err)
		expectedMapping[appearancesTestSchema], err = rowconv.NewFieldMapping(appearancesTestSchema, destSch, appsToDestMapping)
		assert.Nil(t, err)

		assert.Equal(t, expectedMapping, rss.mapping)
		assert.Equal(t, destSch, rss.destSch)
	})

	t.Run("Test validation", func(t *testing.T) {
		colColl, err := schema.NewColCollection(
			schema.NewColumn("id", 0, types.IntKind, true),
			schema.NewColumn("first", 1, types.StringKind, false),
			schema.NewColumn("last", 2, types.StringKind, false),
			schema.NewColumn("is_married", 3, types.BoolKind, false),
			schema.NewColumn("age", 4, types.IntKind, false),
			schema.NewColumn("rating", 5, types.FloatKind, false),
			schema.NewColumn("uuid", 6, types.UUIDKind, false),
		)
		destSch := schema.UnkeyedSchemaFromCols(colColl)
		rss, err := NewFromDestSchema(destSch)
		assert.Nil(t, err)

		// One more column than we're expecting
		assert.NotNil(t, rss.AddSchema(peopleTestSchema))
	})
}

func TestNewFromSourceSchemas(t *testing.T) {
	t.Run("Test with all schemas", func(t *testing.T) {
		colColl := getAllSchemaColumns(t)
		destSch := schema.UnkeyedSchemaFromCols(colColl)

		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema, appearancesTestSchema)
		assert.Nil(t, err)
		peopleToDestMapping := map[uint64]uint64{
			0: 0,
			1: 1,
			2: 2,
			3: 3,
			4: 4,
			// missing 5 tag in people schema
			6: 5,
			7: 6,
			8: 7,
		}
		episodesToDestMapping := map[uint64]uint64{
			0: 8,
			1: 9,
			2: 10,
			3: 11,
		}
		appsToDestMapping := map[uint64]uint64{
			0: 12,
			1: 13,
			2: 14,
		}

		expectedMapping := make(map[schema.Schema]*rowconv.FieldMapping)
		expectedMapping[peopleTestSchema], err = rowconv.NewFieldMapping(peopleTestSchema, destSch, peopleToDestMapping)
		assert.Nil(t, err)
		expectedMapping[episodesTestSchema], err = rowconv.NewFieldMapping(episodesTestSchema, destSch, episodesToDestMapping)
		assert.Nil(t, err)
		expectedMapping[appearancesTestSchema], err = rowconv.NewFieldMapping(appearancesTestSchema, destSch, appsToDestMapping)
		assert.Nil(t, err)

		assert.Equal(t, expectedMapping, rss.mapping)
		assert.Equal(t, destSch, rss.destSch)
	})

	t.Run("Test with all schemas, different order", func(t *testing.T) {
		colColl, err := schema.NewColCollection(
			schema.NewColumn("character_id", 0, types.IntKind, true),
			schema.NewColumn("episode_id", 1, types.IntKind, true),
			schema.NewColumn("comments", 2, types.StringKind, false),
			schema.NewColumn("id", 3, types.IntKind, true),
			schema.NewColumn("name", 4, types.StringKind, false),
			schema.NewColumn("air_date", 5, types.IntKind, false),
			schema.NewColumn("rating", 6, types.FloatKind, false),
			schema.NewColumn("id", 7, types.IntKind, true),
			schema.NewColumn("first", 8, types.StringKind, false),
			schema.NewColumn("last", 9, types.StringKind, false),
			schema.NewColumn("is_married", 10, types.BoolKind, false),
			schema.NewColumn("age", 11, types.IntKind, false),
			schema.NewColumn("rating", 12, types.FloatKind, false),
			schema.NewColumn("uuid", 13, types.UUIDKind, false),
			schema.NewColumn("num_episodes", 14, types.UintKind, false),
		)
		assert.Nil(t, err)
		destSch := schema.UnkeyedSchemaFromCols(colColl)

		rss, err := NewFromSourceSchemas(appearancesTestSchema, episodesTestSchema, peopleTestSchema)
		assert.Nil(t, err)
		appsToDestMapping := map[uint64]uint64{
			0: 0,
			1: 1,
			2: 2,
		}
		episodesToDestMapping := map[uint64]uint64{
			0: 3,
			1: 4,
			2: 5,
			3: 6,
		}
		peopleToDestMapping := map[uint64]uint64{
			0: 7,
			1: 8,
			2: 9,
			3: 10,
			4: 11,
			// missing 5 tag in people schema
			6: 12,
			7: 13,
			8: 14,
		}

		expectedMapping := make(map[schema.Schema]*rowconv.FieldMapping)
		expectedMapping[peopleTestSchema], err = rowconv.NewFieldMapping(peopleTestSchema, destSch, peopleToDestMapping)
		assert.Nil(t, err)
		expectedMapping[episodesTestSchema], err = rowconv.NewFieldMapping(episodesTestSchema, destSch, episodesToDestMapping)
		assert.Nil(t, err)
		expectedMapping[appearancesTestSchema], err = rowconv.NewFieldMapping(appearancesTestSchema, destSch, appsToDestMapping)
		assert.Nil(t, err)

		assert.Equal(t, expectedMapping, rss.mapping)
		assert.Equal(t, destSch, rss.destSch)
	})
}

func mustGetCol(sch schema.Schema, colName string) schema.Column {
	col, ok := sch.GetAllCols().GetByName(colName)
	if !ok {
		panic("No column " + colName)
	}
	return col
}

func TestNewFromColumns(t *testing.T) {
	t.Run("Test cross product", func(t *testing.T) {
		cols := []schema.Column{
			mustGetCol(episodesTestSchema, "id"),
			mustGetCol(peopleTestSchema, "id"),
			mustGetCol(episodesTestSchema, "name"),
			mustGetCol(peopleTestSchema, "first"),
			mustGetCol(peopleTestSchema, "last"),
		}

		rss, err := NewFromColumns(cols...)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Nil(t, rss.AddColumn(episodesTestSchema, mustGetCol(episodesTestSchema, "id")))
		assert.Nil(t, rss.AddColumn(peopleTestSchema, mustGetCol(peopleTestSchema, "id")))
		assert.Nil(t, rss.AddColumn(episodesTestSchema, mustGetCol(episodesTestSchema, "name")))
		assert.Nil(t, rss.AddColumn(peopleTestSchema, mustGetCol(peopleTestSchema, "first")))
		assert.Nil(t, rss.AddColumn(peopleTestSchema, mustGetCol(peopleTestSchema, "last")))

		peopleToDestMapping := map[uint64]uint64{
			0: 1,
			1: 3,
			2: 4,
		}
		episodesToDestMapping := map[uint64]uint64{
			0: 0,
			1: 2,
		}

		expectedMapping := make(map[schema.Schema]*rowconv.FieldMapping)
		expectedMapping[peopleTestSchema], err = rowconv.NewFieldMapping(peopleTestSchema, rss.destSch, peopleToDestMapping)
		assert.Nil(t, err)
		expectedMapping[episodesTestSchema], err = rowconv.NewFieldMapping(episodesTestSchema, rss.destSch, episodesToDestMapping)
		assert.Nil(t, err)

		assert.Equal(t, expectedMapping, rss.mapping)

		tables := []TableResult{
			{Rows: rs(homer, marge), Schema: peopleTestSchema},
			{Rows: rs(ep1, ep2), Schema: episodesTestSchema},
		}

		expectedResult := rs(
			newResultSetRow(mustGetColVal(ep1, episodeIdTag), mustGetColVal(homer, idTag), mustGetColVal(ep1, epNameTag), mustGetColVal(homer, firstTag), mustGetColVal(homer, lastTag)),
			newResultSetRow(mustGetColVal(ep2, episodeIdTag), mustGetColVal(homer, idTag), mustGetColVal(ep2, epNameTag), mustGetColVal(homer, firstTag), mustGetColVal(homer, lastTag)),
			newResultSetRow(mustGetColVal(ep1, episodeIdTag), mustGetColVal(marge, idTag), mustGetColVal(ep1, epNameTag), mustGetColVal(marge, firstTag), mustGetColVal(marge, lastTag)),
			newResultSetRow(mustGetColVal(ep2, episodeIdTag), mustGetColVal(marge, idTag), mustGetColVal(ep2, epNameTag), mustGetColVal(marge, firstTag), mustGetColVal(marge, lastTag)),
		)

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})
}

// Creates a new row for a result set specified by the given values
func newResultSetRow(colVals ...types.Value) row.Row {

	taggedVals := make(row.TaggedValues)
	cols := make([]schema.Column, len(colVals))
	for i := 0; i < len(colVals); i++ {
		taggedVals[uint64(i)] = colVals[i]
		nomsKind := colVals[i].Kind()
		cols[i] = schema.NewColumn(fmt.Sprintf("%v", i), uint64(i), nomsKind, false)
	}

	collection, err := schema.NewColCollection(cols...)
	if err != nil {
		panic("unexpected error " + err.Error())
	}
	sch := schema.UnkeyedSchemaFromCols(collection)

	return row.New(sch, taggedVals)
}

// Creates a new schema for a result set specified by the given pairs of column names and types. Column names are
// strings, types are NomsKinds.
func newResultSetSchema(colNamesAndTypes ...interface{}) schema.Schema {

	if len(colNamesAndTypes) % 2 != 0 {
		panic("Non-even number of inputs passed to newResultSetSchema")
	}

	cols := make([]schema.Column, len(colNamesAndTypes) / 2)
	for i := 0; i < len(colNamesAndTypes); i += 2 {
		name := colNamesAndTypes[i].(string)
		nomsKind := colNamesAndTypes[i+1].(types.NomsKind)
		cols[i/2] = schema.NewColumn(name, uint64(i/2), nomsKind, false)
	}

	collection, err := schema.NewColCollection(cols...)
	if err != nil {
		panic("unexpected error " + err.Error())
	}
	return schema.UnkeyedSchemaFromCols(collection)
}

// Returns all the columns from all 3 test schemas as a single column collection.
func getAllSchemaColumns(t *testing.T) *schema.ColCollection {
	colColl, err := schema.NewColCollection(
		schema.NewColumn("id", 0, types.IntKind, true),
		schema.NewColumn("first", 1, types.StringKind, false),
		schema.NewColumn("last", 2, types.StringKind, false),
		schema.NewColumn("is_married", 3, types.BoolKind, false),
		schema.NewColumn("age", 4, types.IntKind, false),
		schema.NewColumn("rating", 5, types.FloatKind, false),
		schema.NewColumn("uuid", 6, types.UUIDKind, false),
		schema.NewColumn("num_episodes", 7, types.UintKind, false),
		schema.NewColumn("id", 8, types.IntKind, true),
		schema.NewColumn("name", 9, types.StringKind, false),
		schema.NewColumn("air_date", 10, types.IntKind, false),
		schema.NewColumn("rating", 11, types.FloatKind, false),
		schema.NewColumn("character_id", 12, types.IntKind, true),
		schema.NewColumn("episode_id", 13, types.IntKind, true),
		schema.NewColumn("comments", 14, types.StringKind, false),
	)

	if err != nil {
		assert.FailNow(t, "Unexpected error creating schema: %v", err.Error())
	}
	return colColl
}

func TestNewFromDestSchema(t *testing.T) {
	colColl := getAllSchemaColumns(t)

	t.Run("Untyped schema", func(t *testing.T) {
		expectedSch := untyped.UntypeUnkeySchema(schema.UnkeyedSchemaFromCols(colColl))
		sch, _ := ConcatSchemas(untypedPeopleSch, untypedEpisodesSch, untypedAppearacesSch)
		assert.Equal(t, expectedSch, sch)
	})

	t.Run("Typed schema", func(t *testing.T) {
		expectedSch := schema.UnkeyedSchemaFromCols(colColl)
		sch, _ := ConcatSchemas(peopleTestSchema, episodesTestSchema, appearancesTestSchema)
		assert.Equal(t, expectedSch, sch)
	})

}

func TestCombineRows(t *testing.T) {
	t.Run("Combine all schemas", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema, appearancesTestSchema)
		assert.Nil(t, err)

		r := RowWithSchema{row.New(rss.destSch, nil), rss.destSch}
		r = rss.CombineRows(r, RowWithSchema{homer, peopleTestSchema})
		r = rss.CombineRows(r, RowWithSchema{ep1, episodesTestSchema})
		r = rss.CombineRows(r, RowWithSchema{app1, appearancesTestSchema})

		expectedRow := row.New(rss.destSch, row.TaggedValues{
			0: mustGetColVal(homer, idTag),
			1: mustGetColVal(homer, firstTag),
			2: mustGetColVal(homer, lastTag),
			3: mustGetColVal(homer, isMarriedTag),
			4: mustGetColVal(homer, ageTag),
			5: mustGetColVal(homer, ratingTag),
			//6:  mustGetColVal(homer, uuidTag), // don't write nil values
			//7:  mustGetColVal(homer, numEpisodesTag), // don't write nil values
			8:  mustGetColVal(ep1, episodeIdTag),
			9:  mustGetColVal(ep1, epNameTag),
			10: mustGetColVal(ep1, epAirDateTag),
			11: mustGetColVal(ep1, epRatingTag),
			12: mustGetColVal(app1, appCharacterTag),
			13: mustGetColVal(app1, appEpTag),
			14: mustGetColVal(app1, appCommentsTag),
		})

		assert.Equal(t, expectedRow, r.Row)
	})

	t.Run("Combine all schemas in opposite order", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema, appearancesTestSchema)
		assert.Nil(t, err)

		// combine the rows in the opposite order that their schemas were declared
		r := RowWithSchema{row.New(rss.destSch, nil), rss.destSch}
		r = rss.CombineRows(r, RowWithSchema{app1, appearancesTestSchema})
		r = rss.CombineRows(r, RowWithSchema{ep1, episodesTestSchema})
		r = rss.CombineRows(r, RowWithSchema{homer, peopleTestSchema})

		expectedRow := row.New(rss.destSch, row.TaggedValues{
			0: mustGetColVal(homer, idTag),
			1: mustGetColVal(homer, firstTag),
			2: mustGetColVal(homer, lastTag),
			3: mustGetColVal(homer, isMarriedTag),
			4: mustGetColVal(homer, ageTag),
			5: mustGetColVal(homer, ratingTag),
			//6:  mustGetColVal(homer, uuidTag), // don't write nil values
			//7:  mustGetColVal(homer, numEpisodesTag), // don't write nil values
			8:  mustGetColVal(ep1, episodeIdTag),
			9:  mustGetColVal(ep1, epNameTag),
			10: mustGetColVal(ep1, epAirDateTag),
			11: mustGetColVal(ep1, epRatingTag),
			12: mustGetColVal(app1, appCharacterTag),
			13: mustGetColVal(app1, appEpTag),
			14: mustGetColVal(app1, appCommentsTag),
		})

		assert.Equal(t, expectedRow, r.Row)
	})

	t.Run("Combine one schema", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema)
		assert.Nil(t, err)

		r := RowWithSchema{row.New(rss.destSch, nil), rss.destSch}
		r = rss.CombineRows(r, RowWithSchema{homer, peopleTestSchema})

		expectedRow := row.New(rss.destSch, row.TaggedValues{
			0: mustGetColVal(homer, idTag),
			1: mustGetColVal(homer, firstTag),
			2: mustGetColVal(homer, lastTag),
			3: mustGetColVal(homer, isMarriedTag),
			4: mustGetColVal(homer, ageTag),
			5: mustGetColVal(homer, ratingTag),
		})

		assert.Equal(t, expectedRow, r.Row)
	})
}

func TestCrossProduct(t *testing.T) {
	t.Run("3x2 cross product", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema, appearancesTestSchema)
		assert.Nil(t, err)

		tables := []TableResult{
			{Rows: rs(homer, marge), Schema: peopleTestSchema},
			{Rows: rs(ep1, ep2), Schema: episodesTestSchema},
			{Rows: rs(app1, app2), Schema: appearancesTestSchema},
		}

		resultRow := RowWithSchema{Schema: rss.destSch, Row: row.New(rss.destSch, nil)}
		expectedResult := rs(
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}, RowWithSchema{app1, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}, RowWithSchema{app2, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep2, episodesTestSchema}, RowWithSchema{app1, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep2, episodesTestSchema}, RowWithSchema{app2, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}, RowWithSchema{app1, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}, RowWithSchema{app2, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}, RowWithSchema{ep2, episodesTestSchema}, RowWithSchema{app1, appearancesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}, RowWithSchema{ep2, episodesTestSchema}, RowWithSchema{app2, appearancesTestSchema}).Row,
		)

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("3x1 cross product", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema, appearancesTestSchema)
		assert.Nil(t, err)

		tables := []TableResult{
			{Rows: rs(homer), Schema: peopleTestSchema},
			{Rows: rs(ep1), Schema: episodesTestSchema},
			{Rows: rs(app1), Schema: appearancesTestSchema},
		}

		resultRow := RowWithSchema{Schema: rss.destSch, Row: row.New(rss.destSch, nil)}
		expectedResult := rs(
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}, RowWithSchema{app1, appearancesTestSchema}).Row,
		)

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("2x2 cross product", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema)
		assert.Nil(t, err)

		tables := []TableResult{
			{Rows: rs(homer, marge), Schema: peopleTestSchema},
			{Rows: rs(ep1, ep2), Schema: episodesTestSchema},
		}

		resultRow := RowWithSchema{Schema: rss.destSch, Row: row.New(rss.destSch, nil)}
		expectedResult := rs(
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}, RowWithSchema{ep2, episodesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}, RowWithSchema{ep1, episodesTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}, RowWithSchema{ep2, episodesTestSchema}).Row,
		)

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("1x3 cross product", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema)
		assert.Nil(t, err)

		tables := []TableResult{
			{Rows: rs(homer, marge, bart), Schema: peopleTestSchema},
		}

		resultRow := RowWithSchema{Schema: rss.destSch, Row: row.New(rss.destSch, nil)}
		expectedResult := rs(
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{homer, peopleTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{marge, peopleTestSchema}).Row,
			rss.CombineAllRows(resultRow.Copy(), RowWithSchema{bart, peopleTestSchema}).Row,
		)

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("2x0 cross product", func(t *testing.T) {
		rss, err := NewFromSourceSchemas(peopleTestSchema, episodesTestSchema)
		assert.Nil(t, err)

		tables := []TableResult{
			{Rows: rs(homer, marge), Schema: peopleTestSchema},
			{Rows: rs(), Schema: episodesTestSchema},
		}

		expectedResult := make([]row.Row, 0)

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("nil cross product", func(t *testing.T) {
		rss, err := NewFromSourceSchemas()
		assert.Nil(t, err)

		tables := []TableResult{}
		expectedResult := rs()

		result := rss.CrossProduct(tables)
		assert.Equal(t, expectedResult, result)
	})
}

func mustGetColVal(r row.Row, tag uint64) types.Value {
	value, ok := r.GetColVal(tag)
	if !ok {
		return nil
	}
	return value
}


// TODO: refactor sqltestutil.go to its own package (probably not sql) and export these values.
//  This is all copy-pasted from there

const (
	idTag = iota
	firstTag
	lastTag
	isMarriedTag
	ageTag
	emptyTag
	ratingTag
	uuidTag
	numEpisodesTag
)

const (
	episodeIdTag = iota
	epNameTag
	epAirDateTag
	epRatingTag
)

const (
	appCharacterTag = iota
	appEpTag
	appCommentsTag
)

const (
	homerId = iota
	margeId
	bartId
	lisaId
	moeId
	barneyId
)

var peopleTestSchema = createPeopleTestSchema()
var untypedPeopleSch = untyped.UntypeUnkeySchema(peopleTestSchema)
var peopleTableName = "people"

var episodesTestSchema = createEpisodesTestSchema()
var untypedEpisodesSch = untyped.UntypeUnkeySchema(episodesTestSchema)
var episodesTableName = "episodes"

var appearancesTestSchema = createAppearancesTestSchema()
var untypedAppearacesSch = untyped.UntypeUnkeySchema(appearancesTestSchema)
var appearancesTableName = "appearances"

func createPeopleTestSchema() schema.Schema {
	colColl, _ := schema.NewColCollection(
		schema.NewColumn("id", idTag, types.IntKind, true, schema.NotNullConstraint{}),
		schema.NewColumn("first", firstTag, types.StringKind, false, schema.NotNullConstraint{}),
		schema.NewColumn("last", lastTag, types.StringKind, false, schema.NotNullConstraint{}),
		schema.NewColumn("is_married", isMarriedTag, types.BoolKind, false),
		schema.NewColumn("age", ageTag, types.IntKind, false),
		//		schema.NewColumn("empty", emptyTag, types.IntKind, false),
		schema.NewColumn("rating", ratingTag, types.FloatKind, false),
		schema.NewColumn("uuid", uuidTag, types.UUIDKind, false),
		schema.NewColumn("num_episodes", numEpisodesTag, types.UintKind, false),
	)
	return schema.SchemaFromCols(colColl)
}

func createEpisodesTestSchema() schema.Schema {
	colColl, _ := schema.NewColCollection(
		schema.NewColumn("id", episodeIdTag, types.IntKind, true, schema.NotNullConstraint{}),
		schema.NewColumn("name", epNameTag, types.StringKind, false, schema.NotNullConstraint{}),
		schema.NewColumn("air_date", epAirDateTag, types.IntKind, false),
		schema.NewColumn("rating", epRatingTag, types.FloatKind, false),
	)
	return schema.SchemaFromCols(colColl)
}

func createAppearancesTestSchema() schema.Schema {
	colColl, _ := schema.NewColCollection(
		schema.NewColumn("character_id", appCharacterTag, types.IntKind, true, schema.NotNullConstraint{}),
		schema.NewColumn("episode_id", appEpTag, types.IntKind, true, schema.NotNullConstraint{}),
		schema.NewColumn("comments", appCommentsTag, types.StringKind, false),
	)
	return schema.SchemaFromCols(colColl)
}

func newPeopleRow(id int, first, last string, isMarried bool, age int, rating float32) row.Row {
	vals := row.TaggedValues{
		idTag: types.Int(id),
		firstTag: types.String(first),
		lastTag: types.String(last),
		isMarriedTag: types.Bool(isMarried),
		ageTag: types.Int(age),
		ratingTag: types.Float(rating),
	}

	return row.New(peopleTestSchema, vals)
}

func newEpsRow(id int, name string, airdate int, rating float32) row.Row {
	vals := row.TaggedValues{
		episodeIdTag: types.Int(id),
		epNameTag: types.String(name),
		epAirDateTag: types.Int(airdate),
		epRatingTag: types.Float(rating),
	}

	return row.New(episodesTestSchema, vals)
}

func newAppsRow(charId, epId int, comment string) row.Row {
	vals := row.TaggedValues{
		appCharacterTag: types.Int(charId),
		appEpTag : types.Int(epId),
		appCommentsTag: types.String(comment),
	}

	return row.New(appearancesTestSchema, vals)
}

// 6 characters
var homer = newPeopleRow(homerId, "Homer", "Simpson", true, 40, 8.5)
var marge = newPeopleRow(margeId, "Marge", "Simpson", true, 38, 8)
var bart = newPeopleRow(bartId, "Bart", "Simpson", false, 10, 9)
var lisa = newPeopleRow(lisaId, "Lisa", "Simpson", false, 8, 10)
var moe = newPeopleRow(moeId, "Moe", "Szyslak", false, 48, 6.5)
var barney = newPeopleRow(barneyId, "Barney", "Gumble", false, 40, 4)
var allPeopleRows = rs(homer, marge, bart, lisa, moe, barney)

// Actually the first 4 episodes of the show
var ep1 = newEpsRow(1, "Simpsons Roasting On an Open Fire", 629953200, 8.0)
var ep2 = newEpsRow(2, "Bart the Genius", 632372400, 9.0)
var ep3 = newEpsRow(3, "Homer's Odyssey", 632977200, 7.0)
var ep4 = newEpsRow(4, "There's No Disgrace Like Home", 633582000, 8.5)
var allEpsRows = rs(ep1, ep2, ep3, ep4)

// These are made up, not the actual show data
var app1 = newAppsRow(homerId, 1, "Homer is great in this one")
var app2 = newAppsRow(margeId, 1, "Marge is here too")
var app3 = newAppsRow(homerId, 2, "Homer is great in this one too")
var app4 = newAppsRow(bartId, 2, "This episode is named after Bart")
var app5 = newAppsRow(lisaId, 2, "Lisa is here too")
var app6 = newAppsRow(moeId, 2, "I think there's a prank call scene")
var app7 = newAppsRow(homerId, 3, "Homer is in every episode")
var app8 = newAppsRow(margeId, 3, "Marge shows up a lot too")
var app9 = newAppsRow(lisaId, 3, "Lisa is the best Simpson")
var app10 = newAppsRow(barneyId, 3, "I'm making this all up")
// nobody in episode 4, that one was terrible
var allAppsRows = rs(app1,app2,app3,app4,app5,app6,app7,app8,app9,app10)

// Convenience func to avoid the boilerplate of typing []row.Row{} all the time
func rs(rows... row.Row) []row.Row {
	return rows
}
