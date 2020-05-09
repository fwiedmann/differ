/*
 * MIT License
 *
 * Copyright (c) 2019 Felix Wiedmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package observe

import (
	"testing"
)

func TestNewObserver(t *testing.T) {
	validTestObserverConfig := newFakeObserverConfig()
	type args struct {
		observerKind   Kind
		observerConfig Config
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "WithValidKindDeployment",
			args: args{
				observerKind:   AppV1Deployment,
				observerConfig: validTestObserverConfig,
			},
			wantErr: false,
		},

		{
			name: "WithValidKindDaemonSet",
			args: args{
				observerKind:   AppV1DaemonSet,
				observerConfig: validTestObserverConfig,
			},
			wantErr: false,
		},
		{
			name: "WithValidKindStatefulSet",
			args: args{
				observerKind:   AppV1StatefulSet,
				observerConfig: validTestObserverConfig,
			},
			wantErr: false,
		},
		{
			name: "WithInvalidKind",
			args: args{
				observerKind:   "invalidType",
				observerConfig: validTestObserverConfig,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewObserver(tt.args.observerKind, tt.args.observerConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewObserver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
