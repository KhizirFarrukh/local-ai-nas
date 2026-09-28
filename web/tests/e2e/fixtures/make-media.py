"""Makes the media fixtures of the system tests (S02.8-T02), so they are
generated, not downloaded (testdata/SOURCES.md). Needs PyAV (pip install av
numpy); the files are committed, so the tests themselves need nothing.

clip.webm  VP8, 320x240, 10 fps, 12 s: a box moving across, with a bar that
           grows with the time, keyframe every second (for seeking).
tone.webm  Opus, 12 s of a 440 Hz tone (for seeking in the audio player).
"""
import math
import os
import sys
from fractions import Fraction

import av
import numpy as np

out = sys.argv[1] if len(sys.argv) > 1 else os.path.dirname(os.path.abspath(__file__))
SECONDS, FPS, W, H = 12, 10, 320, 240

with av.open(os.path.join(out, "clip.webm"), "w", format="webm") as c:
    s = c.add_stream("libvpx", rate=FPS)
    s.width, s.height, s.pix_fmt = W, H, "yuv420p"
    s.codec_context.gop_size = FPS
    s.codec_context.bit_rate = 150_000
    for i in range(SECONDS * FPS):
        img = np.full((H, W, 3), 30, dtype=np.uint8)
        x = int((W - 40) * i / (SECONDS * FPS))
        img[100:140, x : x + 40] = (230, 120, 40)
        img[H - 12 : H - 4, 8 : 8 + int((W - 16) * i / (SECONDS * FPS))] = (80, 200, 120)
        frame = av.VideoFrame.from_ndarray(img, format="rgb24")
        frame.pts = i
        frame.time_base = Fraction(1, FPS)
        for p in s.encode(frame):
            c.mux(p)
    for p in s.encode():
        c.mux(p)

with av.open(os.path.join(out, "tone.webm"), "w", format="webm") as c:
    rate = 48000
    s = c.add_stream("libopus", rate=rate)
    s.codec_context.bit_rate = 32_000
    n = 960
    for k in range(SECONDS * rate // n):
        t = (np.arange(n) + k * n) / rate
        pcm = (0.2 * np.sin(2 * math.pi * 440 * t)).astype(np.float32).reshape(1, n)
        frame = av.AudioFrame.from_ndarray(pcm, format="flt", layout="mono")
        frame.sample_rate = rate
        frame.pts = k * n
        frame.time_base = Fraction(1, rate)
        for p in s.encode(frame):
            c.mux(p)
    for p in s.encode():
        c.mux(p)
print("ok")
