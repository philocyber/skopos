package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoginAdmissionRefillsAndBoundsDistributedAttempts(t *testing.T) {
	var admission authAdmission
	now := time.Unix(1, 0)
	for i := 0; i < 5; i++ {
		if !admission.allowLogin("192.0.2.1", now) {
			t.Fatal("normal initial login allowance exhausted early")
		}
	}
	if admission.allowLogin("192.0.2.1", now) {
		t.Fatal("sixth immediate attempt was accepted")
	}
	if !admission.allowLogin("192.0.2.1", now.Add(6*time.Second)) {
		t.Fatal("login did not recover after the retry interval")
	}
	var distributed authAdmission
	for i := 0; i < 20; i++ {
		if !distributed.allowLogin(fmt.Sprintf("192.0.2.%d", i), now) {
			t.Fatal("global burst exhausted early")
		}
	}
	if distributed.allowLogin("198.51.100.1", now) {
		t.Fatal("rotating client addresses bypassed the global limit")
	}
	if !distributed.allowLogin("198.51.100.1", now.Add(3*time.Second)) {
		t.Fatal("global admission did not refill")
	}
}

func TestLoginAdmissionPeerStorageIsBounded(t *testing.T) {
	var admission authAdmission
	now := time.Unix(1, 0)
	admission.peers = make(map[string]authBucket)
	for i := 0; i < maxAuthPeers; i++ {
		admission.peers[fmt.Sprint(i)] = authBucket{tokens: 5, last: now}
	}
	if admission.allowLogin("new peer", now) || len(admission.peers) > maxAuthPeers {
		t.Fatal("active peer table exceeded its budget")
	}
	if !admission.allowLogin("new peer", now.Add(10*time.Minute)) {
		t.Fatal("expired peer entries prevented new clients from logging in")
	}
}

func TestPasswordWorkConcurrencyBudget(t *testing.T) {
	var admission authAdmission
	var workers sync.WaitGroup
	accepted := make(chan bool, 32)
	release := make(chan struct{})
	for i := 0; i < cap(accepted); i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ok := admission.beginPasswordWork()
			accepted <- ok
			if ok {
				<-release
				admission.endPasswordWork()
			}
		}()
	}
	count := 0
	for i := 0; i < cap(accepted); i++ {
		if <-accepted {
			count++
		}
	}
	close(release)
	workers.Wait()
	if count != 2 {
		t.Fatalf("admitted %d simultaneous password jobs, want 2", count)
	}
	if !admission.beginPasswordWork() {
		t.Fatal("completed work did not release admission")
	}
	admission.endPasswordWork()
}

func TestLoginRateLimitIgnoresForwardedIdentity(t *testing.T) {
	s := &Server{m: &Manager{}}
	for i := 0; i < 6; i++ {
		r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"wrong","password":"test"}`))
		r.RemoteAddr = fmt.Sprintf("192.0.2.10:%d", 1000+i)
		r.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i))
		w := httptest.NewRecorder()
		s.authLogin(w, r)
		want := http.StatusUnauthorized
		if i == 5 {
			want = http.StatusTooManyRequests
			if w.Header().Get("Retry-After") == "" {
				t.Fatal("busy response lacks retry advice")
			}
		}
		if w.Code != want {
			t.Fatalf("request %d returned %d, want %d", i, w.Code, want)
		}
	}
}
