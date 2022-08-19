package xml

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type xParser struct {
	buf []byte
	i   int
	sz  int
}

func Parse(data []byte) (*Node, error) {
	for i, c := range data {
		if c == '<' {
			if i > 0 {
				data = data[i:]
			}
			break
		}
	}

	xp := xParser{
		buf: data,
		sz:  len(data) - 1,
	}

	c, err := xp.SkipWS()
	if err != nil {
		return nil, err
	}

	if c != '<' {
		return nil, fmt.Errorf("expected '<', found '%c'", c)
	}

	c, err = xp.ReadByte()
	if err != nil {
		return nil, err
	}

	if c == '?' {
		_, err = xp.SkipString("xml ")
		if err != nil {
			return nil, err
		}

		ok, c, err := xp.Find('?')
		if !ok {
			err = fmt.Errorf("expected '?', found '%c'", c)
		}
		if err != nil {
			return nil, err
		}

		ok, c, err = xp.Find('>')
		if !ok {
			err = fmt.Errorf("expected '>', found '%c'", c)
		}
		if err != nil {
			return nil, err
		}

		c, err = xp.SkipWS()
		if err != nil {
			return nil, err
		}

		if c != '<' {
			return nil, fmt.Errorf("expected '<', found '%c'", c)
		}
	} else {
		xp.UnreadByte()
	}

	nd, ok, err := xp.StartNode()
	if err != nil {
		return nil, err
	}

	if !ok {
		err = xp.EndNode(nd)
	}

	return nd, err
}

func (xp *xParser) ReadByte() (byte, error) {
	if xp.i > xp.sz {
		return 0, io.EOF
	}
	b := xp.buf[xp.i]
	xp.i++
	return b, nil
}

func (xp *xParser) UnreadByte() error {
	if xp.i <= 0 {
		return errors.New("reader.UnreadByte: at beginning of string")
	}
	xp.i--
	return nil
}

func (xp *xParser) StartNode() (*Node, bool, error) {
	var sb strings.Builder

	c, err := xp.SkipWS()
	if err != nil {
		return nil, false, err
	}

	if c == '>' || c == '/' {
		return nil, false, errors.New("missing node name")
	}

	if !isAlpha(c) && c != '_' {
		if c == '!' {
			_, err = xp.SkipString("--")
			if err != nil {
				return nil, false, err
			}

			ok, c, err := xp.Find('-')
			if !ok {
				return nil, false, fmt.Errorf("expected '-', found '%c'", c)
			}
			if err != nil {
				return nil, false, err
			}

			c, err = xp.SkipString("->")
			if err != nil {
				return nil, false, err
			}

			if c != '<' {
				return nil, false, fmt.Errorf("expected '<', found '%c'", c)
			}

			return xp.StartNode()
		}

		return nil, false, fmt.Errorf("expected alpha character, found '%c'", c)
	}

	sb.WriteByte(c)

	nd := new(Node)

	for {
		c, err = xp.ReadByte()
		if err != nil {
			break
		}

		if isWS(c) {
			nd.Name = sb.String()
			break
		} else if c == '>' {
			nd.Name = sb.String()
			return nd, false, nil
		} else if c == '/' {
			c, err = xp.SkipWS()
			if err != nil {
				return nil, false, err
			}
			if c == '>' {
				nd.Name = sb.String()
				return nd, true, nil
			}
			return nd, false, fmt.Errorf("expected '>', found '%c'", c)
		} else if c == ':' {
			nd.NS = sb.String()
			sb.Reset()
		} else if !isNodeName(c) {
			return nd, false, fmt.Errorf("expected alpha character, found '%c'", c)
		} else {
			sb.WriteByte(c)
		}
	}

	c, err = xp.SkipWS()
	if err != nil {
		return nd, false, err
	}

	if c == '>' {
		nd.Name = sb.String()
		return nd, false, nil
	}

	if c == '/' {
		c, err = xp.SkipWS()
		if err != nil {
			return nd, false, err
		}
		if c == '>' {
			nd.Name = sb.String()
			return nd, true, nil
		}

		return nd, false, fmt.Errorf("expected '>', found '%c'", c)
	}

	sb.Reset()
	sb.WriteByte(c)

	for {
		c, err = xp.ReadByte()
		if err != nil {
			break
		}

		if isAttribute(c) {
			sb.WriteByte(c)
			continue
		}

		if isWS(c) {
			c, err = xp.SkipWS()
			if err != nil {
				return nd, false, err
			}
		}

		if c == '=' {
			name := sb.String()
			sb.Reset()

			c, err = xp.SkipWS()
			if err != nil {
				return nd, false, err
			}

			if c == '"' {
				value, err := xp.ReadAttribute()
				if err != nil {
					return nd, false, err
				}

				nd.AddAttribute(name, value)

				c, err = xp.SkipWS()
				if err != nil {
					return nd, false, err
				}

				if c == '>' {
					break
				}

				if c == '/' {
					c, err = xp.SkipWS()
					if err != nil {
						return nd, false, err
					}
					if c == '>' {
						return nd, true, nil
					}

					return nd, false, fmt.Errorf("expected '>', found '%c'", c)
				}

				sb.Reset()
				sb.WriteByte(c)
			} else {
				return nd, false, fmt.Errorf("expected '\"', found '%c'", c)
			}
		} else {
			return nd, false, fmt.Errorf("expected '=', found '%c'", c)
		}
	}

	return nd, false, nil
}

func (xp *xParser) EndNode(nd *Node) error {
	var (
		nodes    []*Node
		sb       strings.Builder
		spaces   string
		txt, nds bool
	)

	c, err := xp.SkipWS()
	if err != nil {
		return err
	}

	for {
		if c == '<' {
			c, err = xp.ReadByte()
			if err != nil {
				return err
			}

			if c == '/' {
				name := nd.Name
				if nd.NS != "" {
					name = fmt.Sprintf("%s:%s", nd.NS, name)
				}
				c, err = xp.SkipString(name)
				if err != nil {
					return err
				}

				if c != '>' {
					return fmt.Errorf("expected '>', found '%c'", c)
				}

				break
			}

			if c == '!' {
				s, err := xp.ReadExclamation()
				if err != nil {
					return err
				}
				if s != "" {
					sb.WriteString(s)
					txt = true
					nd.CDATA = true
				}

				c, err = xp.SkipWS()
				if err != nil {
					return err
				}
				continue
			}

			xp.UnreadByte()
			n, closed, err := xp.StartNode()
			if err != nil {
				return err
			}

			if !closed {
				err = xp.EndNode(n)
				if err != nil {
					return err
				}
			}

			if txt {
				s := n.InlineString()
				sb.WriteString(s)

				c, err = xp.ReadByte()
				if err != nil {
					return err
				}
			} else {
				nds = true
				//n.parent = nd
				nodes = append(nodes, n)

				c, spaces, err = xp.CheckWS()
				if err != nil {
					return err
				}
			}
		} else {
			if nds {
				for _, n := range nodes {
					sb.WriteString(n.InlineString())
				}
				nds = false
			}

			if spaces != "" {
				sb.WriteString(spaces)
				spaces = ""
			}

			sb.WriteByte(c)
			txt = true

			c, err = xp.ReadByte()
			if err != nil {
				return err
			}
		}
	}

	if txt {
		nd.Text = sb.String()
	}

	if nds {
		nd.Nodes = nodes
	}

	return nil
}

func (xp *xParser) SkipWS() (c byte, err error) {
	for {
		c, err = xp.ReadByte()
		if err != nil || !isWS(c) {
			break
		}
	}

	return
}

func (xp *xParser) CheckWS() (byte, string, error) {
	i := xp.i
	for {
		c, err := xp.ReadByte()
		if err != nil {
			return c, "", err
		}
		if !isWS(c) {
			return c, string(xp.buf[i : xp.i-1]), err
		}
	}
}

func (xp *xParser) SkipString(s string) (byte, error) {
	bs := []byte(s)
	for i := 0; i < len(bs); i++ {
		c, err := xp.ReadByte()
		if err != nil {
			return c, err
		}

		if c != bs[i] {
			err = fmt.Errorf("expected '%c', found '%c'", bs[i], c)
			return c, err
		}
	}

	return xp.SkipWS()
}

func (xp *xParser) ReadAttribute() (string, error) {
	i := xp.i

	for {
		c, err := xp.ReadByte()
		if err != nil {
			return "", err
		}

		if c == '"' {
			return string(xp.buf[i : xp.i-1]), nil
		}
	}
}

func (xp *xParser) ReadExclamation() (string, error) {
	var (
		i = xp.i
		c = xp.buf[i]
	)

	switch c {
	case '-':
		if i > xp.sz-5 { // --->
			return "", fmt.Errorf("expected '<!---->', found '%s'", string(xp.buf[i-2:]))
		}

		if xp.buf[i+1] != '-' {
			return "", fmt.Errorf("expected '<!--', found '%s'", string(xp.buf[i-2:i+2]))
		}

		i += 2

		for i < xp.sz {
			c := xp.buf[i]
			if c == '-' && i < xp.sz-4 && xp.buf[i+1] == '-' && xp.buf[i+2] == '>' {
				xp.i = i + 3
				return "", nil
			}
			i++
		}
		return "", errors.New("expected end of '<!---->', did not find '-->'")
	case '[':
		if i > xp.sz-10 { // CDATA[]]>
			return "", fmt.Errorf("expected '<![CDATA[]]>', found '%s'", string(xp.buf[i-2:]))
		}

		if xp.buf[i+1] != 'C' && xp.buf[i+2] != 'D' && xp.buf[i+3] != 'A' && xp.buf[i+4] != 'T' && xp.buf[i+5] != 'A' && xp.buf[i+6] != '[' {
			return "", fmt.Errorf("expected '<![CDATA[', found '%s'", string(xp.buf[i-2:i+7]))
		}

		i += 7
		from := i

		for i < xp.sz {
			c := xp.buf[i]
			if c == ']' && i < xp.sz-4 && xp.buf[i+1] == ']' && xp.buf[i+2] == '>' {
				xp.i = i + 3
				s := string(xp.buf[from:i])
				return s, nil
			}
			i++
		}
		return "", errors.New("expected end of '<![CDATA[]]>', did not find ']]>'")
	}

	return "", fmt.Errorf("expected '<![CDATA[]]>' or <!---->, found '%s'", string(xp.buf[i-2:i+1]))
}

func (xp *xParser) Find(target byte) (bool, byte, error) {
	for {
		c, err := xp.ReadByte()
		if err != nil {
			return false, c, err
		}

		if c == target {
			return true, c, nil
		}
	}
}

func isWS(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t' || c == '\r' || c == '\v' || c == '\f'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isNodeName(c byte) bool {
	return isAlpha(c) || c == '.' || c == '_' || c == '-' || isDigit(c)
}

func isAttribute(c byte) bool {
	return isAlpha(c) || c == ':' || c == '_' || c == '-' || isDigit(c) || c == '.'
}
