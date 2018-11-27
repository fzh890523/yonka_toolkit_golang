package service

import log "github.com/golang/glog"

var defaultSvcMgr = &SvcManager{}

func DefaultSvcMgr() *SvcManager {
	return defaultSvcMgr
}

type SvcManager struct {
	startupHooks  []StartHook
	shutdownHooks []ShutdownHook
}

func (sm *SvcManager) AddSvc(i interface{}) int {
	return sm.AddHook(i)
}

func (sm *SvcManager) AddHook(i interface{}) int {
	cnt := 0

	if sm.AddStartupHook(i) {
		cnt++
	}
	if sm.AddShutdownHook(i) {
		cnt++
	}

	return cnt
}

func (sm *SvcManager) AddStartupHook(i interface{}) bool {
	if h, ok := i.(StartHook); ok {
		sm.startupHooks = append(sm.startupHooks, h)
		return true
	}
	return false
}

func (sm *SvcManager) AddShutdownHook(i interface{}) bool {
	if h, ok := i.(ShutdownHook); ok {
		sm.shutdownHooks = append(sm.shutdownHooks, h)
		return true
	}
	return false
}

func (sm *SvcManager) Start() (err error) {
	for _, h := range sm.startupHooks {
		// first service
		if s, ok := h.(Service); ok {
			log.Infof("start service: %s", s.Name())
			if curErr := h.Start(); curErr != nil {
				log.Errorf("start service (%s) met err (%v)", s.Name(), curErr)
				if err == nil {
					err = curErr
				}
				return
			}
		}
	}

	for _, h := range sm.startupHooks {
		// then other
		if _, ok := h.(Service); !ok {
			log.Infof("start...")
			if curErr := h.Start(); curErr != nil {
				log.Errorf("start ... met err (%v)", curErr)
				if err == nil {
					err = curErr
				}
				return
			}
		}
	}

	return
}

func (sm *SvcManager) shutdown() (err error) {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("recover from panic: (%v)", p)
		}
		log.Flush()
	}()

	log.Infof("call shutdown hooks, totally %d", len(sm.shutdownHooks))

	for _, h := range sm.shutdownHooks {
		// first service
		if s, ok := h.(Service); ok {
			log.Infof("shutdown service: %s", s.Name())
			if curErr := h.Shutdown(); curErr != nil {
				log.Errorf("shutdown service (%s) met err (%v)", s.Name(), curErr)
				if err == nil {
					err = curErr
				}
			}
		}
	}

	for _, h := range sm.shutdownHooks {
		// then other
		if _, ok := h.(Service); !ok {
			log.Infof("shutdown non-service...")
			if curErr := h.Shutdown(); curErr != nil {
				log.Errorf("start ... met err (%v)", curErr)
				if err == nil {
					err = curErr
				}
			}
		}
	}

	return
}

type Service interface {
	Name() string
}

type StartHook interface {
	Start() (err error)
}

type ShutdownHook interface {
	Shutdown() (err error)
}
