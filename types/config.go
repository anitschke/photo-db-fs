package types

import (
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	MountPoint string        `json:"mountPoint,omitempty"`
	DB         DB            `json:"db"`
	LogLevel   string        `json:"logLevel,omitempty"`
	Queries    []QueryConfig `json:"queries,omitempty"`
}

type DB struct {
	Type   string `json:"type,omitempty"`
	Source string `json:"source,omitempty"`
}

type QueryConfig struct {
	Name     string         `json:"name,omitempty"`
	Selector SelectorConfig `json:"selector"`
}

type SelectorPropertyMap map[string]SelectorProperty

type SelectorConfig struct {
	Type       string              `json:"type,omitempty"`
	Properties SelectorPropertyMap `json:"properties,omitempty"`
}

type SelectorProperty struct {
	Number    float64          `json:"number,omitempty"`
	Strings   []string         `json:"strings,omitempty"`
	String    string           `json:"string,omitempty"`
	Selectors []SelectorConfig `json:"selectors,omitempty"`
	Selector  *SelectorConfig  `json:"selector,omitempty"`
}

// ConfigToQuery takes a QueryConfig and transforms it into a NamedQuery.
//
// Ideally we would just directly deserialize the JSON into a NamedQuery object,
// however the fact that the NamedQuery has a Selector which is an interface
// makes this a little more complicated. Go json.Unmarshal doesn't support
// deserializing JSON into the correct interface type because for a given blob
// of JSON json.Unmarshal doesn't know what struct that fulfils that interface
// it should deserialize it into. I could do this by making Query implement the
// json.Unmarshaler interface, but then all the selectors that can have children
// selectors also need to implement the json.Unmarshaler interface. So it just
// seemed easier in the long run to declare a few custom struct that can handle
// any selector AND that json.Unmarshal can handle for me. Then I just need to
// implement that transform to the actual query. Not sure if this all payed off,
// perhaps I should just suck it up and implement json.Unmarshaler. But it
// works, so I am going to leave it as is for now.
func ConfigToQuery(config QueryConfig) (NamedQuery, error) {
	s, err := configToSelector(config.Selector)
	if err != nil {
		return NamedQuery{}, fmt.Errorf("error parsing config %q: %w", config.Name, err)
	}
	return NamedQuery{
		Name: config.Name,
		Query: Query{
			Selector: s,
		},
	}, nil
}

func ConfigsToQueries(configs []QueryConfig) ([]NamedQuery, error) {
	namedQueries := make([]NamedQuery, 0, len(configs))
	for _, c := range configs {
		q, err := ConfigToQuery(c)
		if err != nil {
			return nil, err
		}
		namedQueries = append(namedQueries, q)
	}
	return namedQueries, nil
}

func ValidateQueryConfigs(configs []QueryConfig) error {
	names := make(map[string]struct{}, len(configs))
	for _, c := range configs {
		_, ok := names[c.Name]
		if ok {
			return fmt.Errorf("query configs must have unique names, the name %q is used for more than one query", c.Name)
		}
		// TODO validate that c.Name is a valid file name
		// https://github.com/anitschke/photo-db-fs/issues/1
	}
	return nil
}

func configToSelector(config SelectorConfig) (Selector, error) {
	switch t := strings.ToLower(config.Type); t {
	case "hastag": // cspell:disable-line
		return configToHasTag(config)
	case "hasrating": // cspell:disable-line
		return configToHasRatting(config)
	case "and":
		return configToAnd(config)
	case "or":
		return configToOr(config)
	case "difference":
		return configToDifference(config)
	default:
		return nil, fmt.Errorf("invalid selector type %q", config.Type)
	}
}

func configToHasTag(config SelectorConfig) (Selector, error) {
	var s HasTag
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "tag":
			s.Tag.Path = p.Strings
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToHasRatting(config SelectorConfig) (Selector, error) {
	var s HasRating
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "operator":
			s.Operator = RelationalOperator(p.String)
			if err := s.Operator.Validate(); err != nil {
				return nil, err
			}
		case "rating":
			s.Rating = p.Number
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToAnd(config SelectorConfig) (Selector, error) {
	var s And
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "operands":
			s.Operands = make([]Selector, 0, len(p.Selectors))
			for _, opConfig := range p.Selectors {
				op, err := configToSelector(opConfig)
				if err != nil {
					return nil, err
				}
				s.Operands = append(s.Operands, op)
			}
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToOr(config SelectorConfig) (Selector, error) {
	var s Or
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "operands":
			s.Operands = make([]Selector, 0, len(p.Selectors))
			for _, opConfig := range p.Selectors {
				op, err := configToSelector(opConfig)
				if err != nil {
					return nil, err
				}
				s.Operands = append(s.Operands, op)
			}
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}

func configToDifference(config SelectorConfig) (Selector, error) {
	var s Difference
	for name, p := range config.Properties {
		switch n := strings.ToLower(name); n {
		case "starting":
			if p.Selector == nil {
				return nil, errors.New("unspecified starting selector")
			}
			start, err := configToSelector(*p.Selector)
			if err != nil {
				return nil, err
			}
			s.Starting = start
		case "excluding":
			if p.Selector == nil {
				return nil, errors.New("unspecified excluding selector")
			}
			exclude, err := configToSelector(*p.Selector)
			if err != nil {
				return nil, err
			}
			s.Excluding = exclude
		default:
			return nil, fmt.Errorf("invalid property %q", name)
		}
	}
	return s, nil
}
