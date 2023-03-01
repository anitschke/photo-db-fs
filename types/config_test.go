package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryToSelector(t *testing.T) {
	type testData struct {
		name     string
		config   QueryConfig
		expQuery NamedQuery
	}

	td := []testData{
		{
			name: "SimpleHasTag",
			config: QueryConfig{
				Name: "myTagQuery",
				Selector: SelectorConfig{
					Type: "hasTag",
					Properties: SelectorPropertyMap{
						"tag": SelectorProperty{
							Strings: []string{"People", "John Doe"},
						},
					},
				},
			},
			expQuery: NamedQuery{
				Name: "myTagQuery",
				Query: Query{
					Selector: HasTag{
						Tag: Tag{
							Path: []string{"People", "John Doe"},
						},
					},
				},
			},
		},
		{
			name: "SimpleAnd",
			config: QueryConfig{
				Name: "myAndQuery",
				Selector: SelectorConfig{
					Type: "and",
					Properties: SelectorPropertyMap{
						"operands": SelectorProperty{
							Selectors: []SelectorConfig{
								{
									Type: "hasTag",
									Properties: SelectorPropertyMap{
										"tag": SelectorProperty{
											Strings: []string{"People", "Jane Doe"},
										},
									},
								},
								{
									Type: "hasTag",
									Properties: SelectorPropertyMap{
										"tag": SelectorProperty{
											Strings: []string{"Activity", "Hiking"},
										},
									},
								},
							},
						},
					},
				},
			},
			expQuery: NamedQuery{
				Name: "myAndQuery",
				Query: Query{
					Selector: And{
						Operands: []Selector{
							HasTag{
								Tag: Tag{
									Path: []string{"People", "Jane Doe"},
								},
							},
							HasTag{
								Tag: Tag{
									Path: []string{"Activity", "Hiking"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SimpleOr",
			config: QueryConfig{
				Name: "myOrQuery",
				Selector: SelectorConfig{
					Type: "or",
					Properties: SelectorPropertyMap{
						"operands": SelectorProperty{
							Selectors: []SelectorConfig{
								{
									Type: "hasTag",
									Properties: SelectorPropertyMap{
										"tag": SelectorProperty{
											Strings: []string{"People", "John Doe"},
										},
									},
								},
								{
									Type: "hasTag",
									Properties: SelectorPropertyMap{
										"tag": SelectorProperty{
											Strings: []string{"People", "Jane Doe"},
										},
									},
								},
							},
						},
					},
				},
			},
			expQuery: NamedQuery{
				Name: "myOrQuery",
				Query: Query{
					Selector: Or{
						Operands: []Selector{
							HasTag{
								Tag: Tag{
									Path: []string{"People", "John Doe"},
								},
							},
							HasTag{
								Tag: Tag{
									Path: []string{"People", "Jane Doe"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SimpleDifference",
			config: QueryConfig{
				Name: "myDifferenceQuery",
				Selector: SelectorConfig{
					Type: "difference",
					Properties: SelectorPropertyMap{
						"starting": SelectorProperty{
							Selector: &SelectorConfig{
								Type: "hasTag",
								Properties: SelectorPropertyMap{
									"tag": SelectorProperty{
										Strings: []string{"People", "John Doe"},
									},
								},
							},
						},
						"excluding": SelectorProperty{
							Selector: &SelectorConfig{
								Type: "hasTag",
								Properties: SelectorPropertyMap{
									"tag": SelectorProperty{
										Strings: []string{"People", "Jane Doe"},
									},
								},
							},
						},
					},
				},
			},
			expQuery: NamedQuery{
				Name: "myDifferenceQuery",
				Query: Query{
					Selector: Difference{
						Starting: HasTag{
							Tag: Tag{
								Path: []string{"People", "John Doe"},
							},
						},
						Excluding: HasTag{
							Tag: Tag{
								Path: []string{"People", "Jane Doe"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range td {
		t.Run(tt.name, func(t *testing.T) {
			actQuery, err := ConfigToQuery(tt.config)
			assert.NoError(t, err)
			assert.Equal(t, actQuery, tt.expQuery)
		})
	}
}

func TestConfigQueryJsonDecode(t *testing.T) {
	// We save config files in JSON format. This is just a quick and dirty test
	// to make sure the serialization deserialization works correctly by doing a
	// round trip.

	config := Config{
		MountPoint: "/srv/photo-db-fs",
		DB: DB{
			Type:   "digikam-sqlite",
			Source: "/srv/digikam-db",
		},
		LogLevel: "debug",
		Queries: []QueryConfig{
			{
				Name: "myTagQuery",
				Selector: SelectorConfig{
					Type: "hasTag",
					Properties: SelectorPropertyMap{
						"tag": SelectorProperty{
							Strings: []string{"People", "John Doe"},
						},
					},
				},
			},
			{
				Name: "myDifferenceQuery",
				Selector: SelectorConfig{
					Type: "difference",
					Properties: SelectorPropertyMap{
						"starting": SelectorProperty{
							Selector: &SelectorConfig{
								Type: "hasTag",
								Properties: SelectorPropertyMap{
									"tag": SelectorProperty{
										Strings: []string{"People", "John Doe"},
									},
								},
							},
						},
						"excluding": SelectorProperty{
							Selector: &SelectorConfig{
								Type: "hasTag",
								Properties: SelectorPropertyMap{
									"tag": SelectorProperty{
										Strings: []string{"People", "Jane Doe"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	actMarshal, err := json.MarshalIndent(config, "", "    ")
	assert.NoError(t, err)

	expMarshal := `{
    "mountPoint": "/srv/photo-db-fs",
    "db": {
        "type": "digikam-sqlite",
        "source": "/srv/digikam-db"
    },
    "logLevel": "debug",
    "queries": [
        {
            "name": "myTagQuery",
            "selector": {
                "type": "hasTag",
                "properties": {
                    "tag": {
                        "strings": [
                            "People",
                            "John Doe"
                        ]
                    }
                }
            }
        },
        {
            "name": "myDifferenceQuery",
            "selector": {
                "type": "difference",
                "properties": {
                    "excluding": {
                        "selector": {
                            "type": "hasTag",
                            "properties": {
                                "tag": {
                                    "strings": [
                                        "People",
                                        "Jane Doe"
                                    ]
                                }
                            }
                        }
                    },
                    "starting": {
                        "selector": {
                            "type": "hasTag",
                            "properties": {
                                "tag": {
                                    "strings": [
                                        "People",
                                        "John Doe"
                                    ]
                                }
                            }
                        }
                    }
                }
            }
        }
    ]
}`
	assert.Equal(t, string(actMarshal), expMarshal)

	var actUnmarshal Config
	err = json.Unmarshal(actMarshal, &actUnmarshal)
	assert.NoError(t, err)
	assert.Equal(t, actUnmarshal, config)
}
