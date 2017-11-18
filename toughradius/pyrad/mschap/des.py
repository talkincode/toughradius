# This file is part of 'NTLM Authorization Proxy Server'
# Copyright 2001 Dmitry A. Rozmanov <dima@xenon.spb.ru>
#
# NTLM APS is free software; you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation; either version 2 of the License, or
# (at your option) any later version.
#
# NTLM APS is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with the sofware; see the file COPYING. If not, write to the
# Free Software Foundation, Inc.,
# 59 Temple Place, Suite 330, Boston, MA 02111-1307, USA.
#

import des_c, utils

#---------------------------------------------------------------------
class DES:

    des_c_obj = None

    #-----------------------------------------------------------------
    def __init__(self, key_str):
        ""
        k = str_to_key56(key_str)
        k = key56_to_key64(k)
        key_str = utils.lst2str(k)
        self.des_c_obj = des_c.DES(key_str)

    #-----------------------------------------------------------------
    def encrypt(self, plain_text):
        ""
        return self.des_c_obj.encrypt(plain_text)

    #-----------------------------------------------------------------
    def decrypt(self, crypted_text):
        ""
        return self.des_c_obj.decrypt(crypted_text)

#---------------------------------------------------------------------
#Some Helpers
#---------------------------------------------------------------------

DESException = 'DESException'

#---------------------------------------------------------------------
def str_to_key56(key_str):
    ""
    if type(key_str) != type(''):
        #rise DESException, 'ERROR. Wrong key type.'
        pass
    if len(key_str) < 7:
        key_str = key_str + '\000\000\000\000\000\000\000'[:(7 - len(key_str))]
    key_56 = []
    for i in key_str[:7]: key_56.append(ord(i))

    return key_56

#---------------------------------------------------------------------
def key56_to_key64(key_56):
    ""
    key = []
    for i in range(8): key.append(0)

    key[0] = key_56[0];
    key[1] = ((key_56[0] << 7) & 0xFF) | (key_56[1] >> 1);
    key[2] = ((key_56[1] << 6) & 0xFF) | (key_56[2] >> 2);
    key[3] = ((key_56[2] << 5) & 0xFF) | (key_56[3] >> 3);
    key[4] = ((key_56[3] << 4) & 0xFF) | (key_56[4] >> 4);
    key[5] = ((key_56[4] << 3) & 0xFF) | (key_56[5] >> 5);
    key[6] = ((key_56[5] << 2) & 0xFF) | (key_56[6] >> 6);
    key[7] =  (key_56[6] << 1) & 0xFF;

    key = set_key_odd_parity(key)

    return key

#---------------------------------------------------------------------
def set_key_odd_parity(key):
    ""
    for i in range(len(key)):
        for k in range(7):
            bit = 0
            t = key[i] >> k
            bit = (t ^ bit) & 0x1
        key[i] = (key[i] & 0xFE) | bit

    return key
