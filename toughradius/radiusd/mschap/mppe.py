import mschap
import hashlib
import random

SHSpad1 = \
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" + \
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" + \
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" + \
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"

SHSpad2 = \
    "\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2" + \
    "\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2" + \
    "\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2" + \
    "\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2\xf2"

Magic1 = \
    "\x54\x68\x69\x73\x20\x69\x73\x20\x74" + \
    "\x68\x65\x20\x4d\x50\x50\x45\x20\x4d" + \
    "\x61\x73\x74\x65\x72\x20\x4b\x65\x79"

Magic2 = \
    "\x4f\x6e\x20\x74\x68\x65\x20\x63\x6c\x69" + \
    "\x65\x6e\x74\x20\x73\x69\x64\x65\x2c\x20" + \
    "\x74\x68\x69\x73\x20\x69\x73\x20\x74\x68" + \
    "\x65\x20\x73\x65\x6e\x64\x20\x6b\x65\x79" + \
    "\x3b\x20\x6f\x6e\x20\x74\x68\x65\x20\x73" + \
    "\x65\x72\x76\x65\x72\x20\x73\x69\x64\x65" + \
    "\x2c\x20\x69\x74\x20\x69\x73\x20\x74\x68" + \
    "\x65\x20\x72\x65\x63\x65\x69\x76\x65\x20" + \
    "\x6b\x65\x79\x2e"

Magic3 = \
    "\x4f\x6e\x20\x74\x68\x65\x20\x63\x6c\x69" + \
    "\x65\x6e\x74\x20\x73\x69\x64\x65\x2c\x20" + \
    "\x74\x68\x69\x73\x20\x69\x73\x20\x74\x68" + \
    "\x65\x20\x72\x65\x63\x65\x69\x76\x65\x20" + \
    "\x6b\x65\x79\x3b\x20\x6f\x6e\x20\x74\x68" + \
    "\x65\x20\x73\x65\x72\x76\x65\x72\x20\x73" + \
    "\x69\x64\x65\x2c\x20\x69\x74\x20\x69\x73" + \
    "\x20\x74\x68\x65\x20\x73\x65\x6e\x64\x20" + \
    "\x6b\x65\x79\x2e"


def mppe_chap2_gen_keys(password, nt_response):
    """
3.3.  Generating 128-bit Session Keys

   When used in conjunction with MS-CHAP-2 authentication, the initial
   MPPE session keys are derived from the peer's Windows NT password.

   The first step is to obfuscate the peer's password using
   NtPasswordHash() function as described in [8].

      NtPasswordHash(Password, PasswordHash)

   The first 16 octets of the result are then hashed again using the MD4
   algorithm.

      PasswordHashHash = md4(PasswordHash)

   The first 16 octets of this second hash are used together with the
   NT-Response field from the MS-CHAP-2 Response packet [8] as the basis
   for the master session key:

      GetMasterKey(PasswordHashHash, NtResponse, MasterKey)

   Once the master key has been generated, it is used to derive two
   128-bit master session keys, one for sending and one for receiving:

GetAsymmetricStartKey(MasterKey, MasterSendKey, 16, TRUE, TRUE)
GetAsymmetricStartKey(MasterKey, MasterReceiveKey, 16, FALSE, TRUE)

   The master session keys are never used to encrypt or decrypt data;
   they are only used in the derivation of transient session keys.  The
   initial transient session keys are obtained by calling the function
   GetNewKeyFromSHA() (described in [3]):

GetNewKeyFromSHA(MasterSendKey, MasterSendKey, 16, SendSessionKey)
GetNewKeyFromSHA(MasterReceiveKey, MasterReceiveKey, 16,
                                                ReceiveSessionKey)

   Finally, the RC4 tables are initialized using the new session keys:

      rc4_key(SendRC4key, 16, SendSessionKey)
      rc4_key(ReceiveRC4key, 16, ReceiveSessionKey)
    """
    password_hash = mschap.nt_password_hash(password)
    password_hash_hash = mschap.hash_nt_password_hash(password_hash)
    master_key = get_master_key(password_hash_hash, nt_response)
    master_send_key = get_asymetric_start_key(master_key, 16, True, True)
    master_recv_key = get_asymetric_start_key(master_key, 16, False, True)
    return master_send_key, master_recv_key


def get_master_key(password_hash_hash, nt_response):
    """
   GetMasterKey(
   IN  16-octet  PasswordHashHash,
   IN  24-octet  NTResponse,
   OUT 16-octet  MasterKey )
   {
      20-octet Digest

      ZeroMemory(Digest, sizeof(Digest));

      /*
       * SHSInit(), SHSUpdate() and SHSFinal()
       * are an implementation of the Secure Hash Standard [7].
       */

      SHSInit(Context);
      SHSUpdate(Context, PasswordHashHash, 16);
      SHSUpdate(Context, NTResponse, 24);
      SHSUpdate(Context, Magic1, 27);
      SHSFinal(Context, Digest);

      MoveMemory(MasterKey, Digest, 16);
   }

    """
    sha_hash = hashlib.sha1()
    sha_hash.update(password_hash_hash)
    sha_hash.update(nt_response)
    sha_hash.update(Magic1)
    return sha_hash.digest()[:16]


def get_asymetric_start_key(master_key, session_key_length, is_send, is_server):
    """

VOID
   GetAsymetricStartKey(
   IN   16-octet      MasterKey,
   OUT  8-to-16 octet SessionKey,
   IN   INTEGER       SessionKeyLength,
   IN   BOOLEAN       IsSend,
   IN   BOOLEAN       IsServer )
   {

      20-octet Digest;

      ZeroMemory(Digest, 20);

      if (IsSend) {
         if (IsServer) {
            s = Magic3
         } else {
            s = Magic2
         }
      } else {
         if (IsServer) {

            s = Magic2
         } else {
            s = Magic3
         }
      }

      /*
       * SHSInit(), SHSUpdate() and SHSFinal()
       * are an implementation of the Secure Hash Standard [7].
       */

      SHSInit(Context);
      SHSUpdate(Context, MasterKey, 16);
      SHSUpdate(Context, SHSpad1, 40);
      SHSUpdate(Context, s, 84);
      SHSUpdate(Context, SHSpad2, 40);
      SHSFinal(Context, Digest);

      MoveMemory(SessionKey, Digest, SessionKeyLength);
   }
    """
    if is_send:
        if is_server:
            s = Magic3
        else:
            s = Magic2
    else:
        if is_server:
            s = Magic2
        else:
            s = Magic3
    sha_hash = hashlib.sha1()
    sha_hash.update(master_key)
    sha_hash.update(SHSpad1)
    sha_hash.update(s)
    sha_hash.update(SHSpad2)
    return sha_hash.digest()[:session_key_length]


def create_plain_text(key):
    key_len = len(key)
    while (len(key) + 1) % 16: key += "\000"
    return chr(key_len) + key


def create_salts():
    send_salt = create_salt()
    recv_salt = create_salt()
    while send_salt == recv_salt: recv_salt = create_salt()
    return (send_salt, recv_salt)


def create_salt():
    return chr(128 + random.randrange(0, 128)) + chr(random.randrange(0, 256))

def gen_radius_encrypt_keys(send_key, recv_key, secret, request_authenticator):
    send_salt, recv_salt = create_salts()
    _send_key = send_salt + radius_encrypt_keys(
        create_plain_text(send_key),
        secret,
        request_authenticator,
        send_salt
    )
    _recv_key = recv_salt + radius_encrypt_keys(
        create_plain_text(recv_key),
        secret,
        request_authenticator,
        recv_salt
    )

    return _send_key, _recv_key


def radius_encrypt_keys(plain_text, secret, request_authenticator, salt):
    """
  Construct a plaintext version of the String field by concate-
         nating the Key-Length and Key sub-fields.  If necessary, pad
         the resulting string until its length (in octets) is an even
         multiple of 16.  It is recommended that zero octets (0x00) be
         used for padding.  Call this plaintext P.

         Call the shared secret S, the pseudo-random 128-bit Request
         Authenticator (from the corresponding Access-Request packet) R,
         and the contents of the Salt field A.  Break P into 16 octet
         chunks p(1), p(2)...p(i), where i = len(P)/16.  Call the
         ciphertext blocks c(1), c(2)...c(i) and the final ciphertext C.
         Intermediate values b(1), b(2)...c(i) are required.  Encryption
         is performed in the following manner ('+' indicates
         concatenation):

      b(1) = MD5(S + R + A)    c(1) = p(1) xor b(1)   C = c(1)
      b(2) = MD5(S + c(1))     c(2) = p(2) xor b(2)   C = C + c(2)
                  .                      .
                  .                      .
                  .                      .
      b(i) = MD5(S + c(i-1))   c(i) = p(i) xor b(i)   C = C + c(i)

      The   resulting   encrypted   String   field    will    contain
      c(1)+c(2)+...+c(i).
    """
    i = len(plain_text) / 16
    b = hashlib.new("md5", secret + request_authenticator + salt).digest()
    c = xor(plain_text[:16], b)
    result = c
    for x in range(1, i):
        b = hashlib.new("md5", secret + c).digest()
        c = xor(plain_text[x * 16:(x + 1) * 16], b)
        result += c
    return result


def xor(str1, str2):
    return ''.join(map(lambda s1, s2: chr(ord(s1) ^ ord(s2)), str1, str2))
