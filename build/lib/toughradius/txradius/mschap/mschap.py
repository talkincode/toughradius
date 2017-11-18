#!/usr/bin/env python
# coding=utf-8


import des
import md4
import hashlib
import utils


def challenge_hash(peer_challenge, authenticator_challenge, username):
    """ChallengeHash"""
    sha_hash = hashlib.sha1()
    sha_hash.update(peer_challenge)
    sha_hash.update(authenticator_challenge)
    sha_hash.update(username)
    return sha_hash.digest()[:8]


def nt_password_hash(passwd):
    """NtPasswordHash"""
    pw = utils.str2unicode(passwd)
    md4_context = md4.new()
    md4_context.update(pw)
    return md4_context.digest()


def hash_nt_password_hash(password_hash):
    """HashNtPasswordHash"""
    md4_context = md4.new()
    md4_context.update(password_hash)
    return md4_context.digest()


def generate_nt_response_mschap(challenge, password):
    password_hash = nt_password_hash(password)
    return challenge_response(challenge, password_hash)


def generate_nt_response_mschap2(authenticator_challenge, peer_challenge, username, password):
    """GenerateNTResponse"""
    challenge = challenge_hash(peer_challenge, authenticator_challenge, username)
    password_hash = nt_password_hash(password)
    return challenge_response(challenge, password_hash)



def challenge_response(challenge, password_hash):
    """ChallengeResponse"""
    zpassword_hash = password_hash.ljust(21, '\0')

    response = ""
    des_obj = des.DES(zpassword_hash[0:7])
    response += des_obj.encrypt(challenge)

    des_obj = des.DES(zpassword_hash[7:14])
    response += des_obj.encrypt(challenge)

    des_obj = des.DES(zpassword_hash[14:21])
    response += des_obj.encrypt(challenge)
    return response


def generate_authenticator_response(password, nt_response, peer_challenge, authenticator_challenge, username):
    """GenerateAuthenticatorResponse"""
    Magic1 = "\x4D\x61\x67\x69\x63\x20\x73\x65\x72\x76\x65\x72\x20\x74\x6F\x20\x63\x6C\x69\x65\x6E\x74\x20\x73\x69\x67\x6E\x69\x6E\x67\x20\x63\x6F\x6E\x73\x74\x61\x6E\x74"
    Magic2 = "\x50\x61\x64\x20\x74\x6F\x20\x6D\x61\x6B\x65\x20\x69\x74\x20\x64\x6F\x20\x6D\x6F\x72\x65\x20\x74\x68\x61\x6E\x20\x6F\x6E\x65\x20\x69\x74\x65\x72\x61\x74\x69\x6F\x6E"
    password_hash = nt_password_hash(password)
    password_hash_hash = hash_nt_password_hash(password_hash)

    sha_hash = hashlib.sha1()
    sha_hash.update(password_hash_hash)
    sha_hash.update(nt_response)
    sha_hash.update(Magic1)
    digest = sha_hash.digest()

    challenge = challenge_hash(peer_challenge, authenticator_challenge, username)

    sha_hash = hashlib.sha1()
    sha_hash.update(digest)
    sha_hash.update(challenge)
    sha_hash.update(Magic2)
    digest = sha_hash.digest()

    return "\x01S=" + convert_to_hex_string(digest)


def check_authenticator_response(password, nt_response, peer_challenge, authenticator_challenge, user_name, received_response):
    """CheckAuthenticatorResponse"""
    my_resppnse = generate_authenticator_response(password, nt_response, peer_challenge, authenticator_challenge, user_name)

    return my_resppnse == received_response

def convert_to_hex_string(string):
    hex_str = ""
    for c in string:
        hex_tmp = hex(ord(c))[2:]
        if len(hex_tmp) == 1:
            hex_tmp = "0" + hex_tmp
        hex_str += hex_tmp
    return hex_str.upper()





def lm_password_hash(password):

    ucase_password = password.upper()[:14]
    while len(ucase_password) < 14:
        ucase_password += "\0"
    password_hash = des_hash(ucase_password[:7])
    password_hash += des_hash(ucase_password[7:])
    return password_hash


def des_hash(clear):
    """DesEncrypt"""
    des_obj = des.DES(clear)
    return des_obj.encrypt(r"KGS!@#$%")

