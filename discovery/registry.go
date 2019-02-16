package discovery

import "time"

/*
	fqn -> nsregistry
	-	namespace -> vsregistry
		-	version -> route
			-	route -round robin-> []address
*/

type (
	Registry struct {
		// fqn -> nsregistry
		routes    map[string]*nsregistry
		interests []string
	}

	nsregistry struct {
		// namespace -> versionregistry
		versions map[string]*versionregistry
	}

	versionregistry struct {
		// [1.0, 1.1, 2.0]
		routes []*route
		index  int
	}

	route struct {
		address *address
		seen    int64
	}
)

func NewRegistry(interests ...string) *Registry {
	routes := make(map[string]*nsregistry)
	return &Registry{routes, interests}
}

func (r *Registry) Update(a *address) {
	if len(r.interests) > 0 {
		for _, i := range r.interests {
			if a.FQN == i {
				r.update(a)
			}
		}
	} else {
		r.update(a)
	}
}

func (r *Registry) update(a *address) {
	registry, ok := r.routes[a.FQN]

	if ok {
		registry.Update(a)
	} else {
		vs := make(map[string]*versionregistry)
		registry := &nsregistry{vs}
		registry.Update(a)

		r.routes[a.FQN] = registry
	}
}

func (r *Registry) Find(fqn, namespace, version string, maxAge int) *address {
	if fqn == "" {
		fqn = "*"
	}

	registry, ok := r.routes[fqn]

	if namespace == "" {
		namespace = "*"
	}

	if version == "" {
		version = "*"
	}

	if ok {
		return registry.Find(namespace, version, maxAge)
	}

	return nil
}

// --- namespace registry ---

func (n *nsregistry) Update(a *address) {
	sub, ok := n.versions[a.Namespace]

	if ok {
		sub.Update(a)
	} else {
		rs := make([]*route, 0)
		vs := &versionregistry{rs, 1}

		vs.Update(a)
		n.versions[a.Namespace] = vs
	}
}

func (n *nsregistry) Find(namespace, version string, maxAge int) *address {
	sub, ok := n.versions[namespace]

	if ok {
		return sub.Find(version, maxAge)
	}

	return nil
}

// --- version registry ---

func (s *versionregistry) Update(a *address) {
	if len(s.routes) == 0 {
		s.routes = append(s.routes, &route{a, time.Now().Unix()})
	} else {
		c, ok := s.find(a.Version, 0)

		if ok {
			c.seen = time.Now().Unix()
		} else {
			s.routes = append(s.routes, &route{a, time.Now().Unix()})
		}
	}
}

func (s *versionregistry) Find(version string, maxAge int) *address {
	r, ok := s.find(version, 0)

	if ok {
		seconds := time.Duration(maxAge) * time.Second
		if time.Now().Before(time.Unix(r.seen, 0).Add(seconds)) {
			return r.address
		}

		return nil
	}

	return nil
}

func (s *versionregistry) find(version string, attempts int) (*route, bool) {
	if attempts > len(s.routes) {
		return nil, false
	}

	s.incr()

	current := s.routes[s.index-1]

	if version != "*" {
		if current.address.Version != version {
			return s.find(version, attempts+1)
		}
	}

	return current, true
}

func (s *versionregistry) incr() {
	max := len(s.routes)

	s.index = s.index + 1

	if s.index > max {
		s.index = 1
	}
}
