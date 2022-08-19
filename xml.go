package xml

import (
	"fmt"
	"strings"
)

//Node represents XML node with name, optional anames, text and children nodes
type Node struct {
	Name       string
	Nodes      []*Node
	Text       string
	Attributes []Attribute
	CDATA      bool
	NS         string
}

//Attribute XML node attribute
type Attribute struct {
	Name  string
	Value string
}

//Node method finds first child node
func (nd *Node) Node(name string) *Node {
	if nd == nil || len(nd.Nodes) == 0 {
		return nil
	}
	for _, n := range nd.Nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}

//Attribute finds attribute by name
func (nd *Node) Attribute(name string) string {
	if len(nd.Attributes) == 0 {
		return ""
	}
	for _, a := range nd.Attributes {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

//AddAttribute adds attribute to node
func (nd *Node) AddAttribute(name string, value string) {
	attr := Attribute{
		Name:  name,
		Value: value,
	}
	if len(nd.Attributes) == 0 {
		nd.Attributes = []Attribute{attr}
		return
	}
	nd.Attributes = append(nd.Attributes, attr)
}

//Matches compares two nodes
func (nd *Node) Matches(other *Node) (match bool, s string) {
	if nd == nil {
		s = "Left is nil"
		return
	}
	if other == nil {
		s = "Right is nil"
		return
	}

	if nd.Name != other.Name {
		s = fmt.Sprintf("Name mismatch: [ %s ] vs [ %s ]", nd.Name, other.Name)
		return
	}

	if nd.Text != other.Text {
		s = fmt.Sprintf("Text mismatch [ %s ]: %s", nd.Name, nd.Text)
		return
	}

	var (
		lsize = len(nd.Attributes)
		rsize = len(other.Attributes)
	)

	if lsize != rsize {
		s = fmt.Sprintf("Attribute count mismatch: [ %d ] vs [ %d ]", lsize, rsize)
		return
	}
	if lsize > 0 {
		for i, la := range nd.Attributes {
			ra := other.Attributes[i]
			if la.Name != ra.Name {
				s = fmt.Sprintf("Attribute names mismatch: [ %s ] vs [ %s ]", la.Name, ra.Name)
				return
			}
			if la.Value != ra.Value {
				s = fmt.Sprintf("Attribute mismatch: [ %s ] vs [ %s ]", la.Value, ra.Value)
				return
			}
		}
	}

	lsize = len(nd.Nodes)
	rsize = len(other.Nodes)

	if lsize != rsize {
		s = fmt.Sprintf("Node count mismatch: [ %d ] vs [ %d ]", lsize, rsize)
		return
	}
	if lsize > 0 {
		for i, ln := range nd.Nodes {
			rn := other.Nodes[i]
			match, s = ln.Matches(rn)
			if !match {
				return
			}
		}
	}

	match = true
	return
}

//String method serializes Node into pretty XML string
func (nd *Node) String() string {
	return nd.toString(0)
}

//InlineString method serializes Node into condenced XML string
func (nd *Node) InlineString() string {
	var sb strings.Builder

	sb.WriteByte('<')
	if nd.Name == "!" {
		sb.WriteByte('!')
		sb.WriteByte('-')
		sb.WriteByte('-')
		sb.WriteString(nd.Text)
		sb.WriteByte('-')
		sb.WriteByte('-')
		sb.WriteByte('>')
	} else {
		if nd.NS == "" {
			sb.WriteString(nd.Name)
		} else {
			sb.WriteString(nd.NS)
			sb.WriteByte(':')
			sb.WriteString(nd.Name)
		}

		if len(nd.Attributes) > 0 {
			for _, a := range nd.Attributes {
				sb.WriteByte(' ')
				sb.WriteString(a.Name)
				sb.WriteByte('=')
				sb.WriteByte('"')
				sb.WriteString(a.Value)
				sb.WriteByte('"')
			}
		}
		sb.WriteByte('>')
		if len(nd.Nodes) > 0 {
			for _, n := range nd.Nodes {
				s := n.InlineString()
				sb.WriteString(s)
			}
		}
		if nd.Text != "" {
			if nd.CDATA {
				sb.WriteString("<![CDATA[")
			}
			sb.WriteString(strings.TrimSpace(nd.Text))
			if nd.CDATA {
				sb.WriteString("]]>")
			}
		}
		sb.WriteByte('<')
		sb.WriteByte('/')
		if nd.NS == "" {
			sb.WriteString(nd.Name)
		} else {
			sb.WriteString(nd.NS)
			sb.WriteByte(':')
			sb.WriteString(nd.Name)
		}
		sb.WriteByte('>')
	}

	return sb.String()
}

func (nd *Node) toString(level int) string {
	var sb strings.Builder
	if level > 0 {
		i := 0
		for i <= level {
			sb.WriteByte(' ')
			i++
		}
	}

	sb.WriteString("<")
	if nd.Name == "!" {
		sb.WriteByte('!')
		sb.WriteByte('-')
		sb.WriteByte('-')
		sb.WriteString(nd.Text)
		sb.WriteByte('-')
		sb.WriteByte('-')
		sb.WriteByte('>')
	} else {
		if nd.NS == "" {
			sb.WriteString(nd.Name)
		} else {
			sb.WriteString(nd.NS)
			sb.WriteByte(':')
			sb.WriteString(nd.Name)
		}
		if len(nd.Attributes) > 0 {
			for _, a := range nd.Attributes {
				sb.WriteByte(' ')
				sb.WriteString(a.Name)
				sb.WriteByte('=')
				sb.WriteByte('"')
				sb.WriteString(a.Value)
				sb.WriteByte('"')
			}
		}
		sb.WriteString(">")
		if len(nd.Nodes) > 0 {
			for _, n := range nd.Nodes {
				sb.WriteByte('\n')
				sb.WriteString(n.toString(level + 1))
			}
			sb.WriteByte('\n')
			if level > 0 {
				i := 0
				for i <= level {
					sb.WriteByte(' ')
					i++
				}
			}
		}
		if nd.Text != "" {
			if nd.CDATA {
				sb.WriteString("<![CDATA[")
				sb.WriteString(nd.Text)
				sb.WriteString("]]>")
			} else {
				sb.WriteString(strings.TrimSpace(nd.Text))
			}
		}
		sb.WriteByte('<')
		sb.WriteByte('/')
		if nd.NS == "" {
			sb.WriteString(nd.Name)
		} else {
			sb.WriteString(nd.NS)
			sb.WriteByte(':')
			sb.WriteString(nd.Name)
		}
		sb.WriteByte('>')
	}

	return sb.String()
}
