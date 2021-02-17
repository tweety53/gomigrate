package migration

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

import (
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// MigrationsCollectorInterfaceMock implements MigrationsCollectorInterface
type MigrationsCollectorInterfaceMock struct {
	t minimock.Tester

	funcCollectMigrations          func(dirpath string, current int, target int) (m1 Migrations, err error)
	inspectFuncCollectMigrations   func(dirpath string, current int, target int)
	afterCollectMigrationsCounter  uint64
	beforeCollectMigrationsCounter uint64
	CollectMigrationsMock          mMigrationsCollectorInterfaceMockCollectMigrations
}

// NewMigrationsCollectorInterfaceMock returns a mock for MigrationsCollectorInterface
func NewMigrationsCollectorInterfaceMock(t minimock.Tester) *MigrationsCollectorInterfaceMock {
	m := &MigrationsCollectorInterfaceMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.CollectMigrationsMock = mMigrationsCollectorInterfaceMockCollectMigrations{mock: m}
	m.CollectMigrationsMock.callArgs = []*MigrationsCollectorInterfaceMockCollectMigrationsParams{}

	return m
}

type mMigrationsCollectorInterfaceMockCollectMigrations struct {
	mock               *MigrationsCollectorInterfaceMock
	defaultExpectation *MigrationsCollectorInterfaceMockCollectMigrationsExpectation
	expectations       []*MigrationsCollectorInterfaceMockCollectMigrationsExpectation

	callArgs []*MigrationsCollectorInterfaceMockCollectMigrationsParams
	mutex    sync.RWMutex
}

// MigrationsCollectorInterfaceMockCollectMigrationsExpectation specifies expectation struct of the MigrationsCollectorInterface.CollectMigrations
type MigrationsCollectorInterfaceMockCollectMigrationsExpectation struct {
	mock    *MigrationsCollectorInterfaceMock
	params  *MigrationsCollectorInterfaceMockCollectMigrationsParams
	results *MigrationsCollectorInterfaceMockCollectMigrationsResults
	Counter uint64
}

// MigrationsCollectorInterfaceMockCollectMigrationsParams contains parameters of the MigrationsCollectorInterface.CollectMigrations
type MigrationsCollectorInterfaceMockCollectMigrationsParams struct {
	dirpath string
	current int
	target  int
}

// MigrationsCollectorInterfaceMockCollectMigrationsResults contains results of the MigrationsCollectorInterface.CollectMigrations
type MigrationsCollectorInterfaceMockCollectMigrationsResults struct {
	m1  Migrations
	err error
}

// Expect sets up expected params for MigrationsCollectorInterface.CollectMigrations
func (mmCollectMigrations *mMigrationsCollectorInterfaceMockCollectMigrations) Expect(dirpath string, current int, target int) *mMigrationsCollectorInterfaceMockCollectMigrations {
	if mmCollectMigrations.mock.funcCollectMigrations != nil {
		mmCollectMigrations.mock.t.Fatalf("MigrationsCollectorInterfaceMock.CollectMigrations mock is already set by Set")
	}

	if mmCollectMigrations.defaultExpectation == nil {
		mmCollectMigrations.defaultExpectation = &MigrationsCollectorInterfaceMockCollectMigrationsExpectation{}
	}

	mmCollectMigrations.defaultExpectation.params = &MigrationsCollectorInterfaceMockCollectMigrationsParams{dirpath, current, target}
	for _, e := range mmCollectMigrations.expectations {
		if minimock.Equal(e.params, mmCollectMigrations.defaultExpectation.params) {
			mmCollectMigrations.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmCollectMigrations.defaultExpectation.params)
		}
	}

	return mmCollectMigrations
}

// Inspect accepts an inspector function that has same arguments as the MigrationsCollectorInterface.CollectMigrations
func (mmCollectMigrations *mMigrationsCollectorInterfaceMockCollectMigrations) Inspect(f func(dirpath string, current int, target int)) *mMigrationsCollectorInterfaceMockCollectMigrations {
	if mmCollectMigrations.mock.inspectFuncCollectMigrations != nil {
		mmCollectMigrations.mock.t.Fatalf("Inspect function is already set for MigrationsCollectorInterfaceMock.CollectMigrations")
	}

	mmCollectMigrations.mock.inspectFuncCollectMigrations = f

	return mmCollectMigrations
}

// Return sets up results that will be returned by MigrationsCollectorInterface.CollectMigrations
func (mmCollectMigrations *mMigrationsCollectorInterfaceMockCollectMigrations) Return(m1 Migrations, err error) *MigrationsCollectorInterfaceMock {
	if mmCollectMigrations.mock.funcCollectMigrations != nil {
		mmCollectMigrations.mock.t.Fatalf("MigrationsCollectorInterfaceMock.CollectMigrations mock is already set by Set")
	}

	if mmCollectMigrations.defaultExpectation == nil {
		mmCollectMigrations.defaultExpectation = &MigrationsCollectorInterfaceMockCollectMigrationsExpectation{mock: mmCollectMigrations.mock}
	}
	mmCollectMigrations.defaultExpectation.results = &MigrationsCollectorInterfaceMockCollectMigrationsResults{m1, err}
	return mmCollectMigrations.mock
}

//Set uses given function f to mock the MigrationsCollectorInterface.CollectMigrations method
func (mmCollectMigrations *mMigrationsCollectorInterfaceMockCollectMigrations) Set(f func(dirpath string, current int, target int) (m1 Migrations, err error)) *MigrationsCollectorInterfaceMock {
	if mmCollectMigrations.defaultExpectation != nil {
		mmCollectMigrations.mock.t.Fatalf("Default expectation is already set for the MigrationsCollectorInterface.CollectMigrations method")
	}

	if len(mmCollectMigrations.expectations) > 0 {
		mmCollectMigrations.mock.t.Fatalf("Some expectations are already set for the MigrationsCollectorInterface.CollectMigrations method")
	}

	mmCollectMigrations.mock.funcCollectMigrations = f
	return mmCollectMigrations.mock
}

// When sets expectation for the MigrationsCollectorInterface.CollectMigrations which will trigger the result defined by the following
// Then helper
func (mmCollectMigrations *mMigrationsCollectorInterfaceMockCollectMigrations) When(dirpath string, current int, target int) *MigrationsCollectorInterfaceMockCollectMigrationsExpectation {
	if mmCollectMigrations.mock.funcCollectMigrations != nil {
		mmCollectMigrations.mock.t.Fatalf("MigrationsCollectorInterfaceMock.CollectMigrations mock is already set by Set")
	}

	expectation := &MigrationsCollectorInterfaceMockCollectMigrationsExpectation{
		mock:   mmCollectMigrations.mock,
		params: &MigrationsCollectorInterfaceMockCollectMigrationsParams{dirpath, current, target},
	}
	mmCollectMigrations.expectations = append(mmCollectMigrations.expectations, expectation)
	return expectation
}

// Then sets up MigrationsCollectorInterface.CollectMigrations return parameters for the expectation previously defined by the When method
func (e *MigrationsCollectorInterfaceMockCollectMigrationsExpectation) Then(m1 Migrations, err error) *MigrationsCollectorInterfaceMock {
	e.results = &MigrationsCollectorInterfaceMockCollectMigrationsResults{m1, err}
	return e.mock
}

// CollectMigrations implements MigrationsCollectorInterface
func (mmCollectMigrations *MigrationsCollectorInterfaceMock) CollectMigrations(dirpath string, current int, target int) (m1 Migrations, err error) {
	mm_atomic.AddUint64(&mmCollectMigrations.beforeCollectMigrationsCounter, 1)
	defer mm_atomic.AddUint64(&mmCollectMigrations.afterCollectMigrationsCounter, 1)

	if mmCollectMigrations.inspectFuncCollectMigrations != nil {
		mmCollectMigrations.inspectFuncCollectMigrations(dirpath, current, target)
	}

	mm_params := &MigrationsCollectorInterfaceMockCollectMigrationsParams{dirpath, current, target}

	// Record call args
	mmCollectMigrations.CollectMigrationsMock.mutex.Lock()
	mmCollectMigrations.CollectMigrationsMock.callArgs = append(mmCollectMigrations.CollectMigrationsMock.callArgs, mm_params)
	mmCollectMigrations.CollectMigrationsMock.mutex.Unlock()

	for _, e := range mmCollectMigrations.CollectMigrationsMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.m1, e.results.err
		}
	}

	if mmCollectMigrations.CollectMigrationsMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmCollectMigrations.CollectMigrationsMock.defaultExpectation.Counter, 1)
		mm_want := mmCollectMigrations.CollectMigrationsMock.defaultExpectation.params
		mm_got := MigrationsCollectorInterfaceMockCollectMigrationsParams{dirpath, current, target}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmCollectMigrations.t.Errorf("MigrationsCollectorInterfaceMock.CollectMigrations got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmCollectMigrations.CollectMigrationsMock.defaultExpectation.results
		if mm_results == nil {
			mmCollectMigrations.t.Fatal("No results are set for the MigrationsCollectorInterfaceMock.CollectMigrations")
		}
		return (*mm_results).m1, (*mm_results).err
	}
	if mmCollectMigrations.funcCollectMigrations != nil {
		return mmCollectMigrations.funcCollectMigrations(dirpath, current, target)
	}
	mmCollectMigrations.t.Fatalf("Unexpected call to MigrationsCollectorInterfaceMock.CollectMigrations. %v %v %v", dirpath, current, target)
	return
}

// CollectMigrationsAfterCounter returns a count of finished MigrationsCollectorInterfaceMock.CollectMigrations invocations
func (mmCollectMigrations *MigrationsCollectorInterfaceMock) CollectMigrationsAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmCollectMigrations.afterCollectMigrationsCounter)
}

// CollectMigrationsBeforeCounter returns a count of MigrationsCollectorInterfaceMock.CollectMigrations invocations
func (mmCollectMigrations *MigrationsCollectorInterfaceMock) CollectMigrationsBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmCollectMigrations.beforeCollectMigrationsCounter)
}

// Calls returns a list of arguments used in each call to MigrationsCollectorInterfaceMock.CollectMigrations.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmCollectMigrations *mMigrationsCollectorInterfaceMockCollectMigrations) Calls() []*MigrationsCollectorInterfaceMockCollectMigrationsParams {
	mmCollectMigrations.mutex.RLock()

	argCopy := make([]*MigrationsCollectorInterfaceMockCollectMigrationsParams, len(mmCollectMigrations.callArgs))
	copy(argCopy, mmCollectMigrations.callArgs)

	mmCollectMigrations.mutex.RUnlock()

	return argCopy
}

// MinimockCollectMigrationsDone returns true if the count of the CollectMigrations invocations corresponds
// the number of defined expectations
func (m *MigrationsCollectorInterfaceMock) MinimockCollectMigrationsDone() bool {
	for _, e := range m.CollectMigrationsMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.CollectMigrationsMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterCollectMigrationsCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcCollectMigrations != nil && mm_atomic.LoadUint64(&m.afterCollectMigrationsCounter) < 1 {
		return false
	}
	return true
}

// MinimockCollectMigrationsInspect logs each unmet expectation
func (m *MigrationsCollectorInterfaceMock) MinimockCollectMigrationsInspect() {
	for _, e := range m.CollectMigrationsMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to MigrationsCollectorInterfaceMock.CollectMigrations with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.CollectMigrationsMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterCollectMigrationsCounter) < 1 {
		if m.CollectMigrationsMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to MigrationsCollectorInterfaceMock.CollectMigrations")
		} else {
			m.t.Errorf("Expected call to MigrationsCollectorInterfaceMock.CollectMigrations with params: %#v", *m.CollectMigrationsMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcCollectMigrations != nil && mm_atomic.LoadUint64(&m.afterCollectMigrationsCounter) < 1 {
		m.t.Error("Expected call to MigrationsCollectorInterfaceMock.CollectMigrations")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *MigrationsCollectorInterfaceMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockCollectMigrationsInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *MigrationsCollectorInterfaceMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *MigrationsCollectorInterfaceMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockCollectMigrationsDone()
}
