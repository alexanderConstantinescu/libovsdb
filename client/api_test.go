package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ovn-org/libovsdb/ovsdb"
	"github.com/stretchr/testify/assert"
)

func TestAPIListSimple(t *testing.T) {
	cache := apiTestCache(t)
	lscacheList := []Model{
		&testLogicalSwitch{
			UUID:        aUUID0,
			Name:        "ls0",
			ExternalIds: map[string]string{"foo": "bar"},
		},
		&testLogicalSwitch{
			UUID:        aUUID1,
			Name:        "ls1",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID2,
			Name:        "ls2",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID3,
			Name:        "ls4",
			ExternalIds: map[string]string{"foo": "baz"},
			Ports:       []string{"port0", "port1"},
		},
	}
	lscache := map[string]Model{}
	for i := range lscacheList {
		lscache[lscacheList[i].(*testLogicalSwitch).UUID] = lscacheList[i]
	}
	cache.cache["Logical_Switch"] = &RowCache{cache: lscache}
	cache.cache["Logical_Switch_Port"] = newRowCache() // empty

	test := []struct {
		name       string
		initialCap int
		resultCap  int
		resultLen  int
		content    []Model
		err        bool
	}{
		{
			name:       "full",
			initialCap: 0,
			resultCap:  len(lscache),
			resultLen:  len(lscacheList),
			content:    lscacheList,
			err:        false,
		},
		{
			name:       "single",
			initialCap: 1,
			resultCap:  1,
			resultLen:  1,
			content:    lscacheList[0:0],
			err:        false,
		},
		{
			name:       "multiple",
			initialCap: 2,
			resultCap:  2,
			resultLen:  2,
			content:    lscacheList[0:2],
			err:        false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiList: %s", tt.name), func(t *testing.T) {
			var result []testLogicalSwitch
			if tt.initialCap != 0 {
				result = make([]testLogicalSwitch, tt.initialCap)
			}
			api := newAPI(cache)
			err := api.List(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Lenf(t, result, tt.resultLen, "Length should match expected")
				assert.Equal(t, cap(result), tt.resultCap, "Capability should match expected")
				assert.ElementsMatchf(t, tt.content, tt.content, "Content should match")
			}

		})
	}

	t.Run("ApiList: Error wrong type", func(t *testing.T) {
		var result []string
		api := newAPI(cache)
		err := api.List(&result)
		assert.NotNil(t, err)
	})

	t.Run("ApiList: Type Selection", func(t *testing.T) {
		var result []testLogicalSwitchPort
		api := newAPI(cache)
		err := api.List(&result)
		assert.Nil(t, err)
		assert.Len(t, result, 0, "Should be empty since cache is empty")
	})

	t.Run("ApiList: Empty List", func(t *testing.T) {
		result := []testLogicalSwitch{}
		api := newAPI(cache)
		err := api.List(&result)
		assert.Nil(t, err)
		assert.Len(t, result, len(lscacheList))
	})
}

func TestAPIListPredicate(t *testing.T) {
	cache := apiTestCache(t)
	lscacheList := []Model{
		&testLogicalSwitch{
			UUID:        aUUID0,
			Name:        "ls0",
			ExternalIds: map[string]string{"foo": "bar"},
		},
		&testLogicalSwitch{
			UUID:        aUUID1,
			Name:        "magicLs1",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID2,
			Name:        "ls2",
			ExternalIds: map[string]string{"foo": "baz"},
		},
		&testLogicalSwitch{
			UUID:        aUUID3,
			Name:        "magicLs2",
			ExternalIds: map[string]string{"foo": "baz"},
			Ports:       []string{"port0", "port1"},
		},
	}
	lscache := map[string]Model{}
	for i := range lscacheList {
		lscache[lscacheList[i].(*testLogicalSwitch).UUID] = lscacheList[i]
	}
	cache.cache["Logical_Switch"] = &RowCache{cache: lscache}

	test := []struct {
		name      string
		predicate interface{}
		content   []Model
		err       bool
	}{
		{
			name: "none",
			predicate: func(t *testLogicalSwitch) bool {
				return false
			},
			content: []Model{},
			err:     false,
		},
		{
			name: "all",
			predicate: func(t *testLogicalSwitch) bool {
				return true
			},
			content: lscacheList,
			err:     false,
		},
		{
			name: "nil function must fail",
			err:  true,
		},
		{
			name: "arbitrary condition",
			predicate: func(t *testLogicalSwitch) bool {
				return strings.HasPrefix(t.Name, "magic")
			},
			content: []Model{lscacheList[1], lscacheList[3]},
			err:     false,
		},
		{
			name: "error wrong type",
			predicate: func(t testLogicalSwitch) string {
				return "foo"
			},
			err: true,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiListPredicate: %s", tt.name), func(t *testing.T) {
			var result []testLogicalSwitch
			api := newAPI(cache)
			cond := api.Where(api.ConditionFromFunc(tt.predicate))
			err := cond.List(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				if !assert.Nil(t, err) {
					t.Log(err)
				}
				assert.ElementsMatchf(t, tt.content, tt.content, "Content should match")
			}

		})
	}
}

func TestAPIListFields(t *testing.T) {
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
			Name:        "magiclsp1",
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
			Name:        "magiclsp2",
			ExternalIds: map[string]string{"foo": "baz"},
			Enabled:     []bool{true},
		},
	}
	lspcache := map[string]Model{}
	for i := range lspcacheList {
		lspcache[lspcacheList[i].(*testLogicalSwitchPort).UUID] = lspcacheList[i]
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspcache}

	testObj := testLogicalSwitchPort{}

	test := []struct {
		name    string
		fields  []interface{}
		prepare func(*testLogicalSwitchPort)
		content []Model
		err     bool
	}{
		{
			name:    "empty object must match everything",
			content: lspcacheList,
			err:     false,
		},
		{
			name: "List unique by UUID",
			prepare: func(t *testLogicalSwitchPort) {
				t.UUID = aUUID0
			},
			content: []Model{lspcache[aUUID0]},
			err:     false,
		},
		{
			name: "List unique by Index",
			prepare: func(t *testLogicalSwitchPort) {
				t.Name = "lsp2"
			},
			content: []Model{lspcache[aUUID2]},
			err:     false,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiListFields: %s", tt.name), func(t *testing.T) {
			var result []testLogicalSwitchPort
			// Clean object
			testObj = testLogicalSwitchPort{}
			api := newAPI(cache)
			err := api.Where(api.ConditionFromModel(&testObj)).List(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatchf(t, tt.content, tt.content, "Content should match")
			}

		})
	}

	t.Run("ApiListFields: Wrong table", func(t *testing.T) {
		var result []testLogicalSwitchPort
		api := newAPI(cache)
		obj := testLogicalSwitch{
			UUID: aUUID0,
		}

		err := api.Where(api.ConditionFromModel(&obj)).List(&result)
		assert.NotNil(t, err)
	})
}

func TestConditionFromFunc(t *testing.T) {
	test := []struct {
		name string
		arg  interface{}
		err  bool
	}{
		{
			name: "wrong function must fail",
			arg: func(s string) bool {
				return false
			},
			err: true,
		},
		{
			name: "wrong function must fail2 ",
			arg: func(t *testLogicalSwitch) string {
				return "foo"
			},
			err: true,
		},
		{
			name: "correct func should succeed",
			arg: func(t *testLogicalSwitch) bool {
				return true
			},
			err: false,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ConditionFromFunc: %s", tt.name), func(t *testing.T) {
			cache := apiTestCache(t)
			api := newAPI(cache)
			condition := api.ConditionFromFunc(tt.arg)
			if tt.err {
				assert.IsType(t, &errorConditionFactory{}, condition)
			} else {
				assert.IsType(t, &predicateCondFactory{}, condition)
			}
		})
	}
}

func TestConditionFromModel(t *testing.T) {
	var testObj testLogicalSwitch
	test := []struct {
		name  string
		model Model
		conds []Condition
		err   bool
	}{
		{
			name:  "wrong model must fail",
			model: &struct{ a string }{},
			err:   true,
		},
		{
			name: "wrong condition must fail",
			model: &struct {
				a string `ovs:"_uuid"`
			}{},
			conds: []Condition{{Field: "foo"}},
			err:   true,
		},
		{
			name:  "correct model must succeed",
			model: &testLogicalSwitch{},
			err:   false,
		},
		{
			name:  "correct model with valid condition must succeed",
			model: &testObj,
			conds: []Condition{
				{
					Field:    &testObj.Name,
					Function: ovsdb.ConditionEqual,
					Value:    "foo",
				},
				{
					Field:    &testObj.Ports,
					Function: ovsdb.ConditionIncludes,
					Value:    []string{"foo"},
				},
			},
			err: false,
		},
	}

	for _, tt := range test {
		t.Run(fmt.Sprintf("ConditionFromModel: %s", tt.name), func(t *testing.T) {
			cache := apiTestCache(t)
			api := newAPI(cache)
			condition := api.ConditionFromModel(tt.model, tt.conds...)
			if tt.err {
				assert.IsType(t, &errorConditionFactory{}, condition)
			} else {
				if len(tt.conds) > 0 {
					assert.IsType(t, &explicitCondFactory{}, condition)
				} else {
					assert.IsType(t, &equalityCondFactory{}, condition)
				}

			}
		})
	}
}

func TestAPIGet(t *testing.T) {
	cache := apiTestCache(t)
	lsCacheList := []Model{}
	lspCacheList := []Model{
		&testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp0",
			Type:        "foo",
			ExternalIds: map[string]string{"foo": "bar"},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID3,
			Name:        "lsp1",
			Type:        "bar",
			ExternalIds: map[string]string{"foo": "baz"},
		},
	}
	lsCache := map[string]Model{}
	lspCache := map[string]Model{}
	for i := range lsCacheList {
		lsCache[lsCacheList[i].(*testLogicalSwitch).UUID] = lsCacheList[i]
	}
	for i := range lspCacheList {
		lspCache[lspCacheList[i].(*testLogicalSwitchPort).UUID] = lspCacheList[i]
	}
	cache.cache["Logical_Switch"] = &RowCache{cache: lsCache}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspCache}

	test := []struct {
		name    string
		prepare func(Model)
		result  Model
		err     bool
	}{
		{
			name: "empty",
			prepare: func(m Model) {
			},
			err: true,
		},
		{
			name: "non_existing",
			prepare: func(m Model) {
				m.(*testLogicalSwitchPort).Name = "foo"
			},
			err: true,
		},
		{
			name: "by UUID",
			prepare: func(m Model) {
				m.(*testLogicalSwitchPort).UUID = aUUID3
			},
			result: lspCacheList[1],
			err:    false,
		},
		{
			name: "by name",
			prepare: func(m Model) {
				m.(*testLogicalSwitchPort).Name = "lsp0"
			},
			result: lspCacheList[0],
			err:    false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiGet: %s", tt.name), func(t *testing.T) {
			var result testLogicalSwitchPort
			tt.prepare(&result)
			api := newAPI(cache)
			err := api.Get(&result)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equalf(t, tt.result, &result, "Result should match")
			}
		})
	}
}

func TestAPICreate(t *testing.T) {
	cache := apiTestCache(t)
	lsCacheList := []Model{}
	lspCacheList := []Model{
		&testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp0",
			Type:        "foo",
			ExternalIds: map[string]string{"foo": "bar"},
		},
		&testLogicalSwitchPort{
			UUID:        aUUID3,
			Name:        "lsp1",
			Type:        "bar",
			ExternalIds: map[string]string{"foo": "baz"},
		},
	}
	lsCache := map[string]Model{}
	lspCache := map[string]Model{}
	for i := range lsCacheList {
		lsCache[lsCacheList[i].(*testLogicalSwitch).UUID] = lsCacheList[i]
	}
	for i := range lspCacheList {
		lspCache[lspCacheList[i].(*testLogicalSwitchPort).UUID] = lspCacheList[i]
	}
	cache.cache["Logical_Switch"] = &RowCache{cache: lsCache}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspCache}

	test := []struct {
		name   string
		input  []Model
		result []ovsdb.Operation
		err    bool
	}{
		{
			name:  "empty",
			input: []Model{&testLogicalSwitch{}},
			result: []ovsdb.Operation{{
				Op:       "insert",
				Table:    "Logical_Switch",
				Row:      map[string]interface{}{},
				UUIDName: "",
			}},
			err: false,
		},
		{
			name: "With some values",
			input: []Model{&testLogicalSwitch{
				Name: "foo",
			}},
			result: []ovsdb.Operation{{
				Op:       "insert",
				Table:    "Logical_Switch",
				Row:      map[string]interface{}{"name": "foo"},
				UUIDName: "",
			}},
			err: false,
		},
		{
			name: "With named UUID ",
			input: []Model{&testLogicalSwitch{
				UUID: "foo",
			}},
			result: []ovsdb.Operation{{
				Op:       "insert",
				Table:    "Logical_Switch",
				Row:      map[string]interface{}{},
				UUIDName: "foo",
			}},
			err: false,
		},
		{
			name: "Multiple",
			input: []Model{
				&testLogicalSwitch{
					UUID: "fooUUID",
					Name: "foo",
				},
				&testLogicalSwitch{
					UUID: "barUUID",
					Name: "bar",
				},
			},
			result: []ovsdb.Operation{{
				Op:       "insert",
				Table:    "Logical_Switch",
				Row:      map[string]interface{}{"name": "foo"},
				UUIDName: "fooUUID",
			}, {
				Op:       "insert",
				Table:    "Logical_Switch",
				Row:      map[string]interface{}{"name": "bar"},
				UUIDName: "barUUID",
			}},
			err: false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiCreate: %s", tt.name), func(t *testing.T) {
			api := newAPI(cache)
			op, err := api.Create(tt.input...)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equalf(t, tt.result, op, "ovsdb.Operation should match")
			}
		})
	}
}

func TestAPIMutate(t *testing.T) {
	cache := apiTestCache(t)
	lspCache := map[string]Model{
		aUUID0: &testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			Type:        "someType",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
			Tag:         []int{1},
		},
		aUUID1: &testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "lsp1",
			Type:        "someType",
			ExternalIds: map[string]string{"foo": "baz"},
			Tag:         []int{1},
		},
		aUUID2: &testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			Type:        "someOtherType",
			ExternalIds: map[string]string{"foo": "baz"},
			Tag:         []int{1},
		},
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspCache}

	testObj := testLogicalSwitchPort{}

	test := []struct {
		name      string
		condition func(API) ConditionalAPI
		model     Model
		mutations []Mutation
		init      map[string]Model
		result    []ovsdb.Operation
		err       bool
	}{
		{
			name: "select by UUID addElement to set",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromModel(&testLogicalSwitch{
					UUID: aUUID0,
				}))
			},
			mutations: []Mutation{
				{
					Field:   &testObj.Tag,
					Mutator: ovsdb.MutateOperationInsert,
					Value:   []int{5},
				},
			},
			result: []ovsdb.Operation{
				{
					Op:        opMutate,
					Table:     "Logical_Switch_Port",
					Mutations: []interface{}{[]interface{}{"tag", ovsdb.MutateOperationInsert, testOvsSet(t, []int{5})}},
					Where:     []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID0}}},
				},
			},
			err: false,
		},
		{
			name: "select by name delete element from map",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromModel(&testLogicalSwitchPort{
					Name: "lsp2",
				}))
			},
			mutations: []Mutation{
				{
					Field:   &testObj.ExternalIds,
					Mutator: ovsdb.MutateOperationDelete,
					Value:   []string{"foo"},
				},
			},
			result: []ovsdb.Operation{
				{
					Op:        opMutate,
					Table:     "Logical_Switch_Port",
					Mutations: []interface{}{[]interface{}{"external_ids", ovsdb.MutateOperationDelete, testOvsSet(t, []string{"foo"})}},
					Where:     []ovsdb.Condition{{Column: "name", Function: ovsdb.ConditionEqual, Value: "lsp2"}},
				},
			},
			err: false,
		},
		{
			name: "select single by predicate name insert element in map",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromFunc(func(lsp *testLogicalSwitchPort) bool {
					return lsp.Name == "lsp2"
				}))
			},
			mutations: []Mutation{
				{
					Field:   &testObj.ExternalIds,
					Mutator: ovsdb.MutateOperationInsert,
					Value:   map[string]string{"bar": "baz"},
				},
			},
			result: []ovsdb.Operation{
				{
					Op:        opMutate,
					Table:     "Logical_Switch_Port",
					Mutations: []interface{}{[]interface{}{"external_ids", ovsdb.MutateOperationInsert, testOvsMap(t, map[string]string{"bar": "baz"})}},
					Where:     []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID2}}},
				},
			},
			err: false,
		},
		{
			name: "select many by predicate name insert element in map",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromFunc(func(lsp *testLogicalSwitchPort) bool {
					return lsp.Type == "someType"
				}))
			},
			mutations: []Mutation{
				{
					Field:   &testObj.ExternalIds,
					Mutator: ovsdb.MutateOperationInsert,
					Value:   map[string]string{"bar": "baz"},
				},
			},
			result: []ovsdb.Operation{
				{
					Op:        opMutate,
					Table:     "Logical_Switch_Port",
					Mutations: []interface{}{[]interface{}{"external_ids", ovsdb.MutateOperationInsert, testOvsMap(t, map[string]string{"bar": "baz"})}},
					Where:     []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID0}}},
				},
				{
					Op:        opMutate,
					Table:     "Logical_Switch_Port",
					Mutations: []interface{}{[]interface{}{"external_ids", ovsdb.MutateOperationInsert, testOvsMap(t, map[string]string{"bar": "baz"})}},
					Where:     []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID1}}},
				},
			},
			err: false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiMutate: %s", tt.name), func(t *testing.T) {
			api := newAPI(cache)
			cond := tt.condition(api)
			ops, err := cond.Mutate(&testObj, tt.mutations)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatchf(t, tt.result, ops, "ovsdb.Operations should match")
			}
		})
	}
}

func TestAPIUpdate(t *testing.T) {
	cache := apiTestCache(t)
	lspCache := map[string]Model{
		aUUID0: &testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			Type:        "someType",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
			Tag:         []int{1},
		},
		aUUID1: &testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "lsp1",
			Type:        "someType",
			ExternalIds: map[string]string{"foo": "baz"},
			Tag:         []int{1},
			Enabled:     []bool{true},
		},
		aUUID2: &testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			Type:        "someOtherType",
			ExternalIds: map[string]string{"foo": "baz"},
			Tag:         []int{1},
		},
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspCache}

	testObj := testLogicalSwitchPort{}

	test := []struct {
		name      string
		condition func(API) ConditionalAPI
		prepare   func(t *testLogicalSwitchPort)
		result    []ovsdb.Operation
		err       bool
	}{
		{
			name: "select by UUID change multiple field",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromModel(&testLogicalSwitch{
					UUID: aUUID0,
				}))
			},
			prepare: func(t *testLogicalSwitchPort) {
				t.Type = "somethingElse"
				t.Tag = []int{6}
			},
			result: []ovsdb.Operation{
				{
					Op:    opUpdate,
					Table: "Logical_Switch_Port",
					Row:   map[string]interface{}{"type": "somethingElse", "tag": testOvsSet(t, []int{6})},
					Where: []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID0}}},
				},
			},
			err: false,
		},
		{
			name: "select by index change multiple field",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromModel(&testLogicalSwitchPort{
					Name: "lsp1",
				}))
			},
			prepare: func(t *testLogicalSwitchPort) {
				t.Type = "somethingElse"
				t.Tag = []int{6}
			},
			result: []ovsdb.Operation{
				{
					Op:    opUpdate,
					Table: "Logical_Switch_Port",
					Row:   map[string]interface{}{"type": "somethingElse", "tag": testOvsSet(t, []int{6})},
					Where: []ovsdb.Condition{{Column: "name", Function: ovsdb.ConditionEqual, Value: "lsp1"}},
				},
			},
			err: false,
		},
		{
			name: "select by field change multiple field",
			condition: func(a API) ConditionalAPI {
				t := testLogicalSwitchPort{
					Type:    "sometype",
					Enabled: []bool{true},
				}
				return a.Where(a.ConditionFromModel(&t, Condition{
					Field:    &t.Type,
					Function: ovsdb.ConditionEqual,
					Value:    "sometype",
				}))
			},
			prepare: func(t *testLogicalSwitchPort) {
				t.Tag = []int{6}
			},
			result: []ovsdb.Operation{
				{
					Op:    opUpdate,
					Table: "Logical_Switch_Port",
					Row:   map[string]interface{}{"tag": testOvsSet(t, []int{6})},
					Where: []ovsdb.Condition{{Column: "type", Function: ovsdb.ConditionEqual, Value: "sometype"}},
				},
			},
			err: false,
		},
		{
			name: "select by field inequality change multiple field",
			condition: func(a API) ConditionalAPI {
				t := testLogicalSwitchPort{
					Type:    "sometype",
					Enabled: []bool{true},
				}
				return a.Where(a.ConditionFromModel(&t, Condition{
					Field:    &t.Type,
					Function: ovsdb.ConditionNotEqual,
					Value:    "sometype",
				}))
			},
			prepare: func(t *testLogicalSwitchPort) {
				t.Tag = []int{6}
			},
			result: []ovsdb.Operation{
				{
					Op:    opUpdate,
					Table: "Logical_Switch_Port",
					Row:   map[string]interface{}{"tag": testOvsSet(t, []int{6})},
					Where: []ovsdb.Condition{{Column: "type", Function: ovsdb.ConditionNotEqual, Value: "sometype"}},
				},
			},
			err: false,
		},
		{
			name: "select multiple by predicate change multiple field",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromFunc(func(t *testLogicalSwitchPort) bool {
					return t.Enabled != nil && t.Enabled[0] == true
				}))
			},
			prepare: func(t *testLogicalSwitchPort) {
				t.Type = "somethingElse"
				t.Tag = []int{6}
			},
			result: []ovsdb.Operation{
				{
					Op:    opUpdate,
					Table: "Logical_Switch_Port",
					Row:   map[string]interface{}{"type": "somethingElse", "tag": testOvsSet(t, []int{6})},
					Where: []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID0}}},
				},
				{
					Op:    opUpdate,
					Table: "Logical_Switch_Port",
					Row:   map[string]interface{}{"type": "somethingElse", "tag": testOvsSet(t, []int{6})},
					Where: []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID1}}},
				},
			},
			err: false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiUpdate: %s", tt.name), func(t *testing.T) {
			api := newAPI(cache)
			cond := tt.condition(api)
			// clean test Object
			testObj = testLogicalSwitchPort{}
			tt.prepare(&testObj)
			ops, err := cond.Update(&testObj)
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatchf(t, tt.result, ops, "ovsdb.Operations should match")
			}
		})
	}
}

func TestAPIDelete(t *testing.T) {
	cache := apiTestCache(t)
	lspCache := map[string]Model{
		aUUID0: &testLogicalSwitchPort{
			UUID:        aUUID0,
			Name:        "lsp0",
			Type:        "someType",
			ExternalIds: map[string]string{"foo": "bar"},
			Enabled:     []bool{true},
			Tag:         []int{1},
		},
		aUUID1: &testLogicalSwitchPort{
			UUID:        aUUID1,
			Name:        "lsp1",
			Type:        "someType",
			ExternalIds: map[string]string{"foo": "baz"},
			Tag:         []int{1},
			Enabled:     []bool{true},
		},
		aUUID2: &testLogicalSwitchPort{
			UUID:        aUUID2,
			Name:        "lsp2",
			Type:        "someOtherType",
			ExternalIds: map[string]string{"foo": "baz"},
			Tag:         []int{1},
		},
	}
	cache.cache["Logical_Switch_Port"] = &RowCache{cache: lspCache}

	test := []struct {
		name      string
		condition func(API) ConditionalAPI
		result    []ovsdb.Operation
		err       bool
	}{
		{
			name: "select by UUID",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromModel(&testLogicalSwitch{
					UUID: aUUID0,
				}))
			},
			result: []ovsdb.Operation{
				{
					Op:    opDelete,
					Table: "Logical_Switch",
					Where: []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID0}}},
				},
			},
			err: false,
		},
		{
			name: "select by index",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromModel(&testLogicalSwitchPort{
					Name: "lsp1",
				}))
			},
			result: []ovsdb.Operation{
				{
					Op:    opDelete,
					Table: "Logical_Switch_Port",
					Where: []ovsdb.Condition{{Column: "name", Function: ovsdb.ConditionEqual, Value: "lsp1"}},
				},
			},
			err: false,
		},
		{
			name: "select by field equality",
			condition: func(a API) ConditionalAPI {
				t := testLogicalSwitchPort{
					Enabled: []bool{true},
				}
				return a.Where(a.ConditionFromModel(&t, Condition{
					Field:    &t.Type,
					Function: ovsdb.ConditionEqual,
					Value:    "sometype",
				}))
			},
			result: []ovsdb.Operation{
				{
					Op:    opDelete,
					Table: "Logical_Switch_Port",
					Where: []ovsdb.Condition{{Column: "type", Function: ovsdb.ConditionEqual, Value: "sometype"}},
				},
			},
			err: false,
		},
		{
			name: "select multiple by predicate",
			condition: func(a API) ConditionalAPI {
				return a.Where(a.ConditionFromFunc(func(t *testLogicalSwitchPort) bool {
					return t.Enabled != nil && t.Enabled[0] == true
				}))
			},
			result: []ovsdb.Operation{
				{
					Op:    opDelete,
					Table: "Logical_Switch_Port",
					Where: []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID0}}},
				},
				{
					Op:    opDelete,
					Table: "Logical_Switch_Port",
					Where: []ovsdb.Condition{{Column: "_uuid", Function: ovsdb.ConditionEqual, Value: ovsdb.UUID{GoUUID: aUUID1}}},
				},
			},
			err: false,
		},
	}
	for _, tt := range test {
		t.Run(fmt.Sprintf("ApiDelete: %s", tt.name), func(t *testing.T) {
			api := newAPI(cache)
			cond := tt.condition(api)
			ops, err := cond.Delete()
			if tt.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.ElementsMatchf(t, tt.result, ops, "ovsdb.Operations should match")
			}
		})
	}
}
