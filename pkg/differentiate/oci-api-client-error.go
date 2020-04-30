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

package differentiate

import "fmt"

// ClientAPIError represents a error for calls to a registries API
type ClientAPIError struct {
	message string
	Err     error
}

func newAPIErrorF(err error, format string, a ...interface{}) ClientAPIError {
	return ClientAPIError{Err: err, message: fmt.Sprintf(format, a...)}
}

// Error implements the error interface
func (e ClientAPIError) Error() string {
	return e.message
}

// Unwrap support for wrapping errors
func (e ClientAPIError) Unwrap() error {
	return e.Err
}

func newPermissionsError(err error, format string, a ...interface{}) PermissionsError {
	return PermissionsError{Err: err, message: fmt.Sprintf(format, a...)}
}

// PermissionsError represents a error for http status codes 401 or 403
type PermissionsError struct {
	message string
	Err     error
}

// Error implements the error interface
func (pe PermissionsError) Error() string {
	return pe.message
}

// Unwrap support for wrapping errors
func (pe PermissionsError) Unwrap() error {
	return pe.Err
}
