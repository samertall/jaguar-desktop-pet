#!/usr/bin/env python3
"""Make tile icons from the original generated jaguar frame (no external assets)."""
from PIL import Image, ImageDraw, ImageFont, ImageFilter, ImageEnhance
from pathlib import Path
import math
ROOT=Path(__file__).resolve().parents[1]
OUT=ROOT/'store-assets'; OUT.mkdir(exist_ok=True)
src=Image.open(ROOT/'app'/'assets'/'walk_02.png').convert('RGBA')
a=src.getchannel('A')
# Drop stray ultra-faint fringe pixels; keep only natural figure and its antialiasing
src.putalpha(a.point(lambda p: 0 if p<6 else p))
# Increase localized ambient exposure, preserving panther's black appearance
body=ImageEnhance.Brightness(src).enhance(1.14)

def mk_tile(w,h, label=False):
    # Background: nearly-black slate with a warm center halo for contrast.
    im=Image.new('RGBA',(w,h)); pix=im.load()
    for y in range(h):
        for x in range(w):
            fy=y/max(1,h); fx=x/max(1,w)
            center=math.exp(-(((fx-.5)/.40)**2+((fy-.49)/.45)**2)*1.3)
            base=(15+int(18*center)+int(4*(1-fy)),24+int(15*center),33+int(12*center))
            pix[x,y]=(*base,255)
    d=ImageDraw.Draw(im,'RGBA')
    r=max(1,int(min(w,h)*.39)); cx=w//2; cy=int(h*.47)
    # ring/spot behind the jaguar
    for rr in range(r, max(0,r-7),-1):
        d.ellipse((cx-rr,cy-rr,cx+rr,cy+rr), outline=(244,179,91,int(105*(r-rr+1)/7)),width=1)
    d.ellipse((cx-int(r*.80),cy-int(r*.80),cx+int(r*.80),cy+int(r*.80)),fill=(183,133,72,25))
    # Place whole jaguar slightly above the middle so paws are visible
    target_w=int(w*(.90 if w/h>1.3 else 1.06))
    target_h=int(h*(.70 if w/h>1.3 else .45))
    scale=min(target_w/body.width,target_h/body.height)
    resized=body.resize((max(1,round(body.width*scale)),max(1,round(body.height*scale))),Image.Resampling.LANCZOS)
    x=(w-resized.width)//2; y=round(h*.46-resized.height/2)
    im.alpha_composite(resized,(x,y))
    # app accent line
    d=ImageDraw.Draw(im,'RGBA')
    yy=int(h*.90)
    d.line((int(w*.21),yy,int(w*.79),yy), fill=(226,173,105,150),width=max(1,int(min(w,h)*.008)))
    return im.convert('RGB')
for name,sz in {
    'Square44x44Logo.png':(44,44),'Square150x150Logo.png':(150,150),
    'Square310x310Logo.png':(310,310),'Wide310x150Logo.png':(310,150),
    'StoreLogo.png':(50,50), 'StoreTile300x300.png':(300,300),
}.items():
    mk_tile(*sz).save(OUT/name,optimize=True)
print('Generated tile assets:',[x.name for x in OUT.glob('*.png')])
