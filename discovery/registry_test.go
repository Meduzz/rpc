package discovery

import (
	"fmt"
	"testing"
)

func TestVersionRegistry(t *testing.T) {
	sub := newVersion()
	addr := newAddress("", "", "", "")

	sub.Update(addr)

	if len(sub.routes) == 0 {
		fel("Expected numer of routes to be more than 0")
		t.Fail()
	}

	sub.incr()

	if sub.index == 2 {
		fel("Expected index to increase to 2")
		t.Fail()
	}

	a := sub.Find("*", 15)

	if a == nil {
		fel("(a) Expected to find a version with wildcard")
		t.Fail()
	}

	b := sub.Find("test", 15)

	if b == nil {
		fel("(b) Expected to find a version on excact match")
		t.Fail()
	}

	c := sub.Find("1.0", 15)

	if c != nil {
		fel("(c) Did not expect a match on non existing version")
		t.Fail()
	}

	addr2 := newAddress("", "", "", "1.0")
	sub.Update(addr2)

	d := sub.Find("1.0", 15)

	if d == nil {
		fel("(d) Expected to find a version on excact match")
		t.Fail()
	}

	sub.Update(addr)

	e := sub.Find("test", 15)

	if e == nil {
		fel("(e) Expected to find a version on excact match")
		t.Fail()
	}
}

func TestNsRegistry(t *testing.T) {
	sub := newNs()
	addr := newAddress("", "", "", "")

	sub.Update(addr)

	a := sub.Find("*", "*", 15)

	if a == nil {
		fel("(a) Expected to find a version on wildcard matches")
		t.Fail()
	}

	b := sub.Find("*", "test", 15)

	if b == nil {
		fel("(b) Expected to find a version on excact match")
		t.Fail()
	}

	c := sub.Find("test", "*", 15)

	if c != nil {
		fel("(c) Did not expect to find a version on invalid match")
		t.Fail()
	}

	addr2 := newAddress("", "test", "", "")
	sub.Update(addr2)

	d := sub.Find("test", "test", 15)

	if d == nil {
		fel("(d) Expected to find a version on excact match")
		t.Fail()
	}

	sub.Update(addr)

	e := sub.Find("*", "*", 15)

	if e == nil {
		fel("(e) Expected to find a version on wildcard matches")
		t.Fail()
	}
}

func TestRegistry_WithInterests(t *testing.T) {
	sub := NewRegistry("test", "*")

	addr := newAddress("", "", "", "")
	sub.Update(addr)

	a := sub.Find("*", "*", "*", 15)

	if a == nil {
		fel("(a) Expected to find a version on wildcard matches")
		t.Fail()
	}

	addr2 := newAddress("test", "test", "", "test")
	sub.Update(addr2)

	b := sub.Find("test", "test", "test", 15)

	if b == nil {
		fel("(b) Expected to find a version on excact match")
		t.Fail()
	}

	addr3 := newAddress("asdf", "", "", "")
	sub.Update(addr3)

	c := sub.Find("asdf", "*", "*", 15)

	if c != nil {
		fel("(c) Did not expect to find a version on invalid match")
		t.Fail()
	}

	sub.Update(addr)

	d := sub.Find("", "", "", 15)

	if d == nil {
		fel("(d) Expected to find a version on wildcard matches")
		t.Fail()
	}
}

func TestRegistry_WithoutIntrests(t *testing.T) {
	sub := NewRegistry()

	addr := newAddress("", "", "", "")
	sub.Update(addr)

	a := sub.Find("*", "*", "*", 15)

	if a == nil {
		fel("(a) Expected to find a version on wildcard matches")
		t.Fail()
	}

	addr2 := newAddress("asdf", "", "", "")
	sub.Update(addr2)

	b := sub.Find("asdf", "*", "*", 15)

	if b == nil {
		fel("(b) Expected to find a version on excact match")
		t.Fail()
	}
}

func newVersion() *versionregistry {
	r := make([]*route, 0)
	return &versionregistry{r, 1}
}

func newNs() *nsregistry {
	v := make(map[string]*versionregistry)
	return &nsregistry{v}
}

func newAddress(fqn, namespace, topic, version string) *address {
	if fqn == "" {
		fqn = "*"
	}

	if namespace == "" {
		namespace = "*"
	}

	if topic == "" {
		topic = "test"
	}

	if version == "" {
		version = "test"
	}

	return &address{
		FQN:       fqn,
		Namespace: namespace,
		Topic:     topic,
		Version:   version,
	}
}

func fel(text string, extras ...interface{}) {
	l := fmt.Sprintf(text, extras...)
	fmt.Printf("%s.\n", l)
}
