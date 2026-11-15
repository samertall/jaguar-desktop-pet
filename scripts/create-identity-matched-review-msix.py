#!/usr/bin/env python3
"""Create an identity-matched, UNSIGNED review MSIX without the Windows SDK.

Purpose: allow archive/manifest review on non-Windows systems. The release
package must be independently verified with the official Windows SDK MakeAppx
and Windows App Certification Kit before certification submission.
"""
from pathlib import Path
import base64
import hashlib
import xml.etree.ElementTree as ET
import zipfile

ROOT = Path(__file__).resolve().parents[1]
NAME = 'SamerTallauze.JaguarDesktopPet'
PUBLISHER = 'CN=C83AFC34-5CE3-4E0A-8DCD-9733097BE6AC'
DISPLAY = 'Samer Tallauze'
VERSION = '0.1.0.0'
OUTPUT = ROOT/'build'/'SamerTallauze.JaguarDesktopPet_0.1.0.0_x64_IDENTITY_MATCHED_REVIEW.msix'


def produce():
    manifest = (ROOT/'packaging'/'AppxManifest.template.xml').read_text('utf8')
    manifest = (manifest.replace('__IDENTITY_NAME__', NAME)
                        .replace('__IDENTITY_PUBLISHER__', PUBLISHER)
                        .replace('__PUBLISHER_DISPLAY_NAME__', DISPLAY))
    doc = ET.fromstring(manifest)
    ns = {'f': 'http://schemas.microsoft.com/appx/manifest/foundation/windows10'}
    node = doc.find('f:Identity', ns)
    assert node is not None and node.attrib['Name'] == NAME
    assert node.attrib['Publisher'] == PUBLISHER
    assert node.attrib['Version'] == VERSION
    manifest_bytes = manifest.encode('utf-8')
    (ROOT/'packaging'/'AppxManifest.IDENTITY-MATCHED.xml').write_bytes(manifest_bytes)

    payload = {'AppxManifest.xml': manifest_bytes, 'JaguarPet.exe': (ROOT/'app'/'JaguarPet.exe').read_bytes()}
    for path in sorted((ROOT/'store-assets').glob('*.png')):
        if path.name == 'StoreTile300x300.png':  # Marketing tile, not an app package resource.
            continue
        payload[f'Assets/{path.name}'] = path.read_bytes()
    assert payload['JaguarPet.exe'][:2] == b'MZ'

    # MSIX uncompressed ZIP blocks with per-block SHA256 references.
    lines = ['<?xml version="1.0" encoding="utf-8"?>',
             '<BlockMap xmlns="http://schemas.microsoft.com/appx/2010/blockmap" HashMethod="http://www.w3.org/2001/04/xmlenc#sha256">']
    for name, data in payload.items():
        lines.append(f'  <File Name="{name.replace("/", chr(92))}" Size="{len(data)}" LfhSize="{30 + len(name.encode("utf-8"))}">')
        for i in range(0, len(data), 65536):
            hashed = base64.b64encode(hashlib.sha256(data[i:i+65536]).digest()).decode('ascii')
            lines.append(f'    <Block Hash="{hashed}" />')
        lines.append('  </File>')
    lines.append('</BlockMap>')
    blockmap = ('\n'.join(lines)+'\n').encode('utf-8')
    ct = b'''<?xml version="1.0" encoding="utf-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="exe" ContentType="application/octet-stream" />
  <Default Extension="png" ContentType="image/png" />
  <Override PartName="/AppxManifest.xml" ContentType="application/vnd.ms-appx.manifest+xml" />
  <Override PartName="/AppxBlockMap.xml" ContentType="application/vnd.ms-appx.blockmap+xml" />
</Types>
'''
    OUTPUT.parent.mkdir(exist_ok=True)
    with zipfile.ZipFile(OUTPUT, 'w', zipfile.ZIP_STORED, allowZip64=True) as z:
        for name, data in payload.items():
            z.writestr(name, data, compress_type=zipfile.ZIP_STORED)
        z.writestr('AppxBlockMap.xml', blockmap, compress_type=zipfile.ZIP_STORED)
        z.writestr('[Content_Types].xml', ct, compress_type=zipfile.ZIP_STORED)
    return OUTPUT

if __name__ == '__main__':
    print(produce())
