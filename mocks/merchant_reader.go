// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/libsv/go-p4"
	"sync"
)

// Ensure, that MerchantReaderMock does implement p4.MerchantReader.
// If this is not the case, regenerate this file with moq.
var _ p4.MerchantReader = &MerchantReaderMock{}

// MerchantReaderMock is a mock implementation of p4.MerchantReader.
//
// 	func TestSomethingThatUsesMerchantReader(t *testing.T) {
//
// 		// make and configure a mocked p4.MerchantReader
// 		mockedMerchantReader := &MerchantReaderMock{
// 			OwnerFunc: func(ctx context.Context) (*p4.MerchantData, error) {
// 				panic("mock out the Owner method")
// 			},
// 		}
//
// 		// use mockedMerchantReader in code that requires p4.MerchantReader
// 		// and then make assertions.
//
// 	}
type MerchantReaderMock struct {
	// OwnerFunc mocks the Owner method.
	OwnerFunc func(ctx context.Context) (*p4.MerchantData, error)

	// calls tracks calls to the methods.
	calls struct {
		// Owner holds details about calls to the Owner method.
		Owner []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockOwner sync.RWMutex
}

// Owner calls OwnerFunc.
func (mock *MerchantReaderMock) Owner(ctx context.Context) (*p4.MerchantData, error) {
	if mock.OwnerFunc == nil {
		panic("MerchantReaderMock.OwnerFunc: method is nil but MerchantReader.Owner was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockOwner.Lock()
	mock.calls.Owner = append(mock.calls.Owner, callInfo)
	mock.lockOwner.Unlock()
	return mock.OwnerFunc(ctx)
}

// OwnerCalls gets all the calls that were made to Owner.
// Check the length with:
//     len(mockedMerchantReader.OwnerCalls())
func (mock *MerchantReaderMock) OwnerCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockOwner.RLock()
	calls = mock.calls.Owner
	mock.lockOwner.RUnlock()
	return calls
}
