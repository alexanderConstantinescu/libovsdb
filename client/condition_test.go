package client

import (
	"fmt"
	"testing"

	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/stretchr/testify/assert"
)

func TestEqualityCondFactory(t *testing.T) {
	cache := apiTestCache(t)
	lspcacheList := []Model{
		&testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "lsp1",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			ExternalIds: map[string]string{"unique": "id"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID3,
			Name:        "lsp3",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{true},
		},
	}
	lspcache := map[string]Model{}
	for i := range lspcacheList {
		lspcache[lspcacheList[i].(*testLogicalSwitchPort).UUID] = lspcacheList[i]
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspcache}

	test := []struct {
		name      string
		model     Model
		condition []ovsdb.Condition
		matches   map[Model]bool
		err       bool
	}{
		{
			name:  "by uuid",
			model: &testLogicalSwitchPort{UUID: aUUID0, Name: "different"},
			condition: []ovsdb.Condition{
				{
					Column:   "_uuid",
					Function: ovsdb.ConditionEqual,
					Value:    ovsdb.UUID{GoUUID: aUUID0},
				}},
			matches: map[Model]bool{
				&testLogicalSwitchPort{UUID: aUUID0}:              true,
				&testLogicalSwitchPort{UUID: aUUID1}:              false,
				&testLogicalSwitchPort{UUID: aUUID0, Name: "foo"}: true,
			},
		},
		{
			name:  "by index",
			model: &testLogicalSwitchPort{Name: "lsp1"},
			condition: []ovsdb.Condition{
				{
					Column:   "name",
					Function: ovsdb.ConditionEqual,
					Value:    "lsp1",
				}},
			matches: map[Model]bool{
				&testLogicalSwitchPort{UUID: aUUID1}:               false,
				&testLogicalSwitchPort{UUID: aUUID1, Name: "lsp1"}: true,
				&testLogicalSwitchPort{UUID: aUUID0, Name: "lsp1"}: true,
			},
		},
		{
			name:  "by non index",
			model: &testLogicalSwitchPort{ExternalIds: map[string]string{"foo": "baz"}},
			err:   true,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("Equality Condition: %s", tt.name), func(t *testing.T) {
			cond, err := newEqualityConditionFactory(cache.orm, "Logical_Switch_Port", tt.model)
			assert.Nil(t, err)
			for model, shouldMatch := range tt.matches {
				matches, err := cond.Matches(model)
				if tt.err {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
					assert.Equalf(t, shouldMatch, matches, fmt.Sprintf("Match on model %#+v should be %v", model, shouldMatch))
				}
			}
			generated, err := cond.Generate()
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatch(t, tt.condition, generated)
			}
		})
	}
}

func TestPredicateCondFactory(t *testing.T) {
	cache := apiTestCache(t)
	lspcacheList := []Model{
		&testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "lsp1",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			ExternalIds: map[string]string{"unique": "id"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID3,
			Name:        "lsp3",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{true},
		},
	}
	lspcache := map[string]Model{}
	for i := range lspcacheList {
		lspcache[lspcacheList[i].(*testLogicalSwitchPort).UUID] = lspcacheList[i]
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspcache}

	test := []struct {
		name      string
		predicate interface{}
		condition []ovsdb.Condition
		matches   map[Model]bool
		err       bool
	}{
		{
			name: "simple value comparison",
			predicate: func(lsp *testLogicalSwitchPort) bool {
				return lsp.UUID == aUUID0
			},
			condition: []ovsdb.Condition{
				{
					Column:   "_uuid",
					Function: ovsdb.ConditionEqual,
					Value:    ovsdb.UUID{GoUUID: aUUID0},
				}},
			matches: map[Model]bool{
				&testLogicalSwitchPort{UUID: aUUID0}:              true,
				&testLogicalSwitchPort{UUID: aUUID1}:              false,
				&testLogicalSwitchPort{UUID: aUUID0, Name: "foo"}: true,
			},
		},
		{
			name: "by random field",
			predicate: func(lsp *testLogicalSwitchPort) bool {
				return lsp.Enabled[0] == false
			},
			condition: []ovsdb.Condition{
				{
					Column:   "_uuid",
					Function: ovsdb.ConditionEqual,
					Value:    ovsdb.UUID{GoUUID: aUUID1},
				},
				{
					Column:   "_uuid",
					Function: ovsdb.ConditionEqual,
					Value:    ovsdb.UUID{GoUUID: aUUID2},
				}},
			matches: map[Model]bool{
				&testLogicalSwitchPort{UUID: aUUID1, Enabled: []bool{true}}:  false,
				&testLogicalSwitchPort{UUID: aUUID1, Enabled: []bool{false}}: true,
			},
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("Predicate Condition: %s", tt.name), func(t *testing.T) {
			cond, err := newPredicateConditionFactory("Logical_Switch_Port", cache, tt.predicate)
			assert.Nil(t, err)
			for model, shouldMatch := range tt.matches {
				matches, err := cond.Matches(model)
				if tt.err {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
					assert.Equalf(t, shouldMatch, matches, fmt.Sprintf("Match on model %#+v should be %v", model, shouldMatch))
				}
			}
			generated, err := cond.Generate()
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatch(t, tt.condition, generated)
			}
		})
	}
}

func TestExplicitCondFactory(t *testing.T) {
	cache := apiTestCache(t)
	lspcacheList := []Model{
		&testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "lsp1",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			ExternalIds: map[string]string{"unique": "id"},
			Enabled:     []bool{false},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID3,
			Name:        "lsp3",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{true},
		},
	}
	lspcache := map[string]Model{}
	for i := range lspcacheList {
		lspcache[lspcacheList[i].(*testLogicalSwitchPort).UUID] = lspcacheList[i]
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspcache}

	testObj := &testLogicalSwitchPort{}

	test := []struct {
		name   string
		args   []Condition
		result []ovsdb.Condition
		err    bool
	}{
		{
			name: "inequality comparison",
			args: []Condition{
				{
					Field:    &testObj.Name,
					Function: ovsdb.ConditionNotEqual,
					Value:    "lsp0",
				},
			},
			result: []ovsdb.Condition{
				{
					Column:   "name",
					Function: ovsdb.ConditionNotEqual,
					Value:    "lsp0",
				}},
		},
		{
			name: "map comparison",
			args: []Condition{
				{
					Field:    &testObj.ExternalIds,
					Function: ovsdb.ConditionIncludes,
					Value:    map[string]string{"foo": "baz"},
				},
			},
			result: []ovsdb.Condition{
				{
					Column:   "external_ids",
					Function: ovsdb.ConditionIncludes,
					Value:    testOvsMap(t, map[string]string{"foo": "baz"}),
				}},
		},
		{
			name: "set comparison",
			args: []Condition{
				{
					Field:    &testObj.Enabled,
					Function: ovsdb.ConditionEqual,
					Value:    []bool{true},
				},
			},
			result: []ovsdb.Condition{
				{
					Column:   "enabled",
					Function: ovsdb.ConditionEqual,
					Value:    testOvsSet(t, []bool{true}),
				}},
		},
		{
			name: "multiple conditions",
			args: []Condition{
				{
					Field:    &testObj.Enabled,
					Function: ovsdb.ConditionEqual,
					Value:    []bool{true},
				},
				{
					Field:    &testObj.Name,
					Function: ovsdb.ConditionNotEqual,
					Value:    "foo",
				},
			},
			result: []ovsdb.Condition{
				{
					Column:   "enabled",
					Function: ovsdb.ConditionEqual,
					Value:    testOvsSet(t, []bool{true}),
				},
				{
					Column:   "name",
					Function: ovsdb.ConditionNotEqual,
					Value:    "foo",
				}},
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("Explicit Condition: %s", tt.name), func(t *testing.T) {
			cond, err := newExplicitConditionFactory(cache.orm, "Logical_Switch_Port", testObj, tt.args...)
			assert.Nil(t, err)
			_, err = cond.Matches(testObj)
			assert.NotNilf(t, err, "Explicit conditions should fail to match on cache")
			generated, err := cond.Generate()
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatch(t, tt.result, generated)
			}
		})
	}
}
