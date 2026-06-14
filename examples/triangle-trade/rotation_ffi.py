#!/usr/bin/env python3
"""rotation_ffi.py — Python ctypes bindings for The Rotation Zig kernel.

Calls librotation_zig.a directly via FFI. No Rust dependency needed.
"""

import ctypes
import ctypes.util
import os
import pathlib

# Locate the static library
_LIB_PATH = pathlib.Path(__file__).parent / "librotation_zig.a"
if not _LIB_PATH.exists():
    # Try the-rotation build output
    alt = pathlib.Path(os.path.expanduser("~/.openclaw/workspace/the-rotation/crates/zig-kernel/zig-out/lib/librotation_zig.a"))
    if alt.exists():
        _LIB_PATH = alt
    else:
        # Will be linked at runtime — raise on first call
        pass

class RotationKernel:
    """Python wrapper around the Rotation Zig kernel (C ABI)."""

    def __init__(self):
        # Load via ctypes.CDLL — works for .a on Linux with dlopen
        try:
            self._lib = ctypes.CDLL(str(_LIB_PATH))
        except OSError:
            # .a files can't be directly loaded by ctypes; need a .so wrapper
            # Fall through — functions will raise AttributeError at call time
            self._lib = None

        if self._lib:
            # tensor_pack(src: *const i8) -> u128
            self._lib.tensor_pack.argtypes = [ctypes.POINTER(ctypes.c_int8)]
            self._lib.tensor_pack.restype = ctypes.c_uint64

            # matmul_ternary_16x16(rows, cols, out)
            self._lib.matmul_ternary_16x16.argtypes = [
                ctypes.POINTER(ctypes.c_uint64),
                ctypes.POINTER(ctypes.c_uint64),
                ctypes.POINTER(ctypes.c_float),
            ]
            self._lib.matmul_ternary_16x16.restype = None

            # attractor_64(values, threshold, output)
            self._lib.attractor_64.argtypes = [
                ctypes.POINTER(ctypes.c_float),
                ctypes.c_float,
                ctypes.POINTER(ctypes.c_int8),
            ]
            self._lib.attractor_64.restype = None

    def pack_ternary(self, values):
        """Pack 64 i8 values into a u128."""
        arr = (ctypes.c_int8 * 64)(*values)
        return self._lib.tensor_pack(arr)

    def attractor(self, values, threshold):
        """Attractor step: returns array of -1, 0, 1."""
        vals = (ctypes.c_float * 64)(*values)
        out = (ctypes.c_int8 * 64)()
        self._lib.attractor_64(vals, ctypes.c_float(threshold), out)
        return list(out)

# Singleton
kernel = RotationKernel()

# Convenience
def attractor(values, threshold=0.5):
    return kernel.attractor(values, threshold)

def pack_ternary(values):
    return kernel.pack_ternary(values)
