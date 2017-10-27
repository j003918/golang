// rpn_test
package rpn

import (
	"testing"
)

func Test_Get_RPN_1(t *testing.T) {
	str_exp := "a+b"
	str_rst := "a,b,+"
	if str_rst != Get_RPN(str_exp, ",") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_2(t *testing.T) {
	str_exp := "a*b+c"
	str_rst := "ab*c+"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_3(t *testing.T) {
	str_exp := "a+b*c"
	str_rst := "abc*+"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_4(t *testing.T) {
	str_exp := "a+(b-c)*d"
	str_rst := "abc-d*+"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_5(t *testing.T) {
	str_exp := "(b-c)*d+a"
	str_rst := "bc-d*a+"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_6(t *testing.T) {
	str_exp := "a+d*(b-c)"
	str_rst := "adbc-*+"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_7(t *testing.T) {
	str_exp := "a*(b-c*d)+e-f/g*(h+i*j-k)"
	str_rst := "abcd*-*e+fg/hij*+k-*-"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}

func Test_Get_RPN_8(t *testing.T) {
	str_exp := "(a+b)*(c-d)+(e/f)-(g+h)*i"
	str_rst := "ab+cd-*ef/+gh+i*-"
	if str_rst != Get_RPN(str_exp, "") {
		t.Error("not pass ")
	} else {
		t.Log("pass")
	}
}
