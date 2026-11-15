#!/usr/bin/env python3
"""Verify ZIP + block map + exact Partner Center identity. NOT a Windows runtime test."""
import base64
import hashlib
from pathlib import Path
import zipfile
import xml.etree.ElementTree as ET

ROOT=Path(__file__).resolve().parents[1]
PKG=ROOT/'build'/'SamerTallauze.JaguarDesktopPet_0.1.0.0_x64_IDENTITY_MATCHED_REVIEW.msix'
EXPECTED_NAME='SamerTallauze.JaguarDesktopPet'
EXPECTED_PUBLISHER='CN=C83AFC34-5CE3-4E0A-8DCD-9733097BE6AC'
F='http://schemas.microsoft.com/appx/manifest/foundation/windows10'
BM='http://schemas.microsoft.com/appx/2010/blockmap'
with zipfile.ZipFile(PKG) as z:
    assert z.testzip() is None
    names = set(z.namelist())
    assert {'AppxManifest.xml','JaguarPet.exe','AppxBlockMap.xml','[Content_Types].xml'} <= names
    tree=ET.fromstring(z.read('AppxManifest.xml'))
    identity=tree.find(f'{{{F}}}Identity')
    assert identity is not None
    assert identity.get('Name')==EXPECTED_NAME
    assert identity.get('Publisher')==EXPECTED_PUBLISHER
    assert identity.get('ProcessorArchitecture')=='x64'
    assert identity.get('Version')=='0.1.0.0'
    props=tree.find(f'{{{F}}}Properties')
    assert props is not None and props.find(f'{{{F}}}PublisherDisplayName').text=='Samer Tallauze'
    deps=tree.find(f'{{{F}}}Dependencies')
    assert deps.find(f'{{{F}}}TargetDeviceFamily').get('Name')=='Windows.Desktop'
    assert z.read('JaguarPet.exe')[:2]==b'MZ'
    blocks=ET.fromstring(z.read('AppxBlockMap.xml'))
    checked=set()
    for entry in blocks.findall(f'{{{BM}}}File'):
        name=entry.attrib['Name'].replace('\\','/')
        assert name in names, name
        info=z.getinfo(name)
        assert info.compress_type==zipfile.ZIP_STORED
        assert int(entry.attrib['Size'])==info.file_size
        assert int(entry.attrib['LfhSize'])==30 + len(info.filename.encode('utf8')) + len(info.extra)
        data=z.read(name)
        expected=[base64.b64encode(hashlib.sha256(data[i:i+65536]).digest()).decode('ascii') for i in range(0,len(data),65536)]
        found=[b.attrib['Hash'] for b in entry.findall(f'{{{BM}}}Block')]
        assert expected==found, name
        checked.add(name)
    assert checked==names-{'AppxBlockMap.xml','[Content_Types].xml'}
    print(f'PASS: Identity matched; {len(checked)} payloads; ZIP, XML, block hashes, PE signature, x64 desktop target.')
    print(f'SHA256: {hashlib.sha256(PKG.read_bytes()).hexdigest()}')
    print('NOT VERIFIED: official MakeAppx validation, Windows launch, certification, antivirus reputation.')
