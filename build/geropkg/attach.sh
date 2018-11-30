#!/bin/bash

ROOT=$(cd `dirname $0`; pwd)
${ROOT}/bin/gero --datadir=${ROOT}/data attach

