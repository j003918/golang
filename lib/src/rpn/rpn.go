// rpn project rpn.go
package rpn

import (
	"github.com/j003918/datastruct/stack"
)

func opr_level(opr string) int {
	switch opr {
	case "+", "-":
		return 1
	case "*", "/", "%":
		return 2
	case "(":
		fallthrough
	default:
		return 0
	}
}

func Get_RPN(expression string, split string) string {
	str_rpn := ""
	str_tmp := ""
	ch := ""
	rpn := stack.New()
	opr := stack.New()

	for _, v := range expression {
		ch = string(v)
		if ch == " " {
			continue
		}

		switch ch {
		case "(":
			if str_tmp != "" {
				rpn.Push(str_tmp)
				str_tmp = ""
			}
			opr.Push(ch)
		case ")":
			if str_tmp != "" {
				rpn.Push(str_tmp)
				str_tmp = ""
			}
			str := ""
			for !opr.Empty() {
				str = opr.Pop().(string)
				if str == "(" {
					break
				}
				rpn.Push(str)
			}
		case "+", "-", "*", "%", "/":
			if str_tmp != "" {
				rpn.Push(str_tmp)
				str_tmp = ""
			}

			for !opr.Empty() {
				if opr_level(opr.Top().(string)) >= opr_level(ch) {
					rpn.Push(opr.Pop())
				} else {
					break
				}
			}
			opr.Push(ch)
		default:
			str_tmp += ch
		}
	}
	if str_tmp != "" {
		rpn.Push(str_tmp)
	}

	for !opr.Empty() {
		rpn.Push(opr.Pop())
	}

	for e := rpn.Front(); e != nil; e = e.Next() {
		str_rpn += e.Value.(string) + split
	}

	opr.Clean()
	rpn.Clean()

	return str_rpn[:len(str_rpn)-len(split)]
}
