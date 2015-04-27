import des
import md4
import sha
import utils


def generate_nt_response_mschap(challenge,password):
    """
   NtChallengeResponse(
   IN  8-octet               Challenge,
   IN  0-to-256-unicode-char Password,
   OUT 24-octet              Response )
   {
      NtPasswordHash( Password, giving PasswordHash )
      ChallengeResponse( Challenge, PasswordHash, giving Response )
   }    
    """
    password_hash=nt_password_hash(password)
    return challenge_response(challenge,password_hash)

def generate_lm_response_mschap(challenge,password):
    """
   NtChallengeResponse(
   IN  8-octet               Challenge,
   IN  0-to-256-unicode-char Password,
   OUT 24-octet              Response )
   {
      NtPasswordHash( Password, giving PasswordHash )
      ChallengeResponse( Challenge, PasswordHash, giving Response )
   }    
    """
    password_hash=lm_password_hash(password)
    return challenge_response(challenge,password_hash)

def generate_nt_response_mschap2(authenticator_challenge,peer_challenge,username,password):
    """
   GenerateNTResponse(
   IN  16-octet              AuthenticatorChallenge,
   IN  16-octet              PeerChallenge,

   IN  0-to-256-char         UserName,

   IN  0-to-256-unicode-char Password,
   OUT 24-octet              Response )
   {
      8-octet  Challenge
      16-octet PasswordHash

      ChallengeHash( PeerChallenge, AuthenticatorChallenge, UserName,
                     giving Challenge)

      NtPasswordHash( Password, giving PasswordHash )
      ChallengeResponse( Challenge, PasswordHash, giving Response )
   }
    
    """
    challenge=challenge_hash(peer_challenge,authenticator_challenge,username)
    password_hash=nt_password_hash(password)
    return challenge_response(challenge,password_hash)


def challenge_hash(peer_challenge,authenticator_challenge,username):
    """
   ChallengeHash(
   IN 16-octet               PeerChallenge,
   IN 16-octet               AuthenticatorChallenge,
   IN  0-to-256-char         UserName,
   OUT 8-octet               Challenge
   {

      /*
       * SHAInit(), SHAUpdate() and SHAFinal() functions are an
       * implementation of Secure Hash Algorithm (SHA-1) [11]. These are
       * available in public domain or can be licensed from
       * RSA Data Security, Inc.
       */

      SHAInit(Context)
      SHAUpdate(Context, PeerChallenge, 16)
      SHAUpdate(Context, AuthenticatorChallenge, 16)

      /*
       * Only the user name (as presented by the peer and
       * excluding any prepended domain name)
       * is used as input to SHAUpdate().
       */

      SHAUpdate(Context, UserName, strlen(Username))
      SHAFinal(Context, Digest)
      memcpy(Challenge, Digest, 8)
   }

    
    """
    sha_hash=sha.new()
    sha_hash.update(peer_challenge)
    sha_hash.update(authenticator_challenge)
    sha_hash.update(username)
    return sha_hash.digest()[:8]
    
def nt_password_hash(passwd,pad_to_21_bytes=True):
    """
   NtPasswordHash(
   IN  0-to-256-unicode-char Password,
   OUT 16-octet              PasswordHash )
   {
      /*
       * Use the MD4 algorithm [5] to irreversibly hash Password
       * into PasswordHash.  Only the password is hashed without
       * including any terminating 0.
       */
    """

    # we have to have UNICODE password
    pw = utils.str2unicode(passwd)

    # do MD4 hash
    md4_context = md4.new()
    md4_context.update(pw)

    res = md4_context.digest()

    if pad_to_21_bytes:
        # addig zeros to get 21 bytes string
	res = res + '\000\000\000\000\000'

    return res


def challenge_response(challenge,password_hash):
    """
   ChallengeResponse(
   IN  8-octet  Challenge,
   IN  16-octet PasswordHash,
   OUT 24-octet Response )
   {
      Set ZPasswordHash to PasswordHash zero-padded to 21 octets

      DesEncrypt( Challenge,
                  1st 7-octets of ZPasswordHash,
                  giving 1st 8-octets of Response )

      DesEncrypt( Challenge,
                  2nd 7-octets of ZPasswordHash,
                  giving 2nd 8-octets of Response )

      DesEncrypt( Challenge,
                  3rd 7-octets of ZPasswordHash,
                  giving 3rd 8-octets of Response )
   }
    """
    zpassword_hash=password_hash
#    while len(zpassword_hash)<21:
#	zpassword_hash+="\0"
    
    response=""
    des_obj=des.DES(zpassword_hash[0:7])
    response+=des_obj.encrypt(challenge)

    des_obj=des.DES(zpassword_hash[7:14])
    response+=des_obj.encrypt(challenge)
    
    des_obj=des.DES(zpassword_hash[14:21])
    response+=des_obj.encrypt(challenge)
    return response


def generate_authenticator_response(password,nt_response,peer_challenge,authenticator_challenge,username):
    """
   GenerateAuthenticatorResponse(
   IN  0-to-256-unicode-char Password,
   IN  24-octet              NT-Response,
   IN  16-octet              PeerChallenge,
   IN  16-octet              AuthenticatorChallenge,
   IN  0-to-256-char         UserName,
   OUT 42-octet              AuthenticatorResponse )
   {
      16-octet              PasswordHash
      16-octet              PasswordHashHash
      8-octet               Challenge

      /*
       * "Magic" constants used in response generation
       */

      Magic1[39] =
         {0x4D, 0x61, 0x67, 0x69, 0x63, 0x20, 0x73, 0x65, 0x72, 0x76,
          0x65, 0x72, 0x20, 0x74, 0x6F, 0x20, 0x63, 0x6C, 0x69, 0x65,
          0x6E, 0x74, 0x20, 0x73, 0x69, 0x67, 0x6E, 0x69, 0x6E, 0x67,
          0x20, 0x63, 0x6F, 0x6E, 0x73, 0x74, 0x61, 0x6E, 0x74};

      Magic2[41] =
         {0x50, 0x61, 0x64, 0x20, 0x74, 0x6F, 0x20, 0x6D, 0x61, 0x6B,
          0x65, 0x20, 0x69, 0x74, 0x20, 0x64, 0x6F, 0x20, 0x6D, 0x6F,
          0x72, 0x65, 0x20, 0x74, 0x68, 0x61, 0x6E, 0x20, 0x6F, 0x6E,
          0x65, 0x20, 0x69, 0x74, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6F,
          0x6E};

      /*
       * Hash the password with MD4
       */

      NtPasswordHash( Password, giving PasswordHash )

      /*
       * Now hash the hash
       */

      HashNtPasswordHash( PasswordHash, giving PasswordHashHash)

      SHAInit(Context)
      SHAUpdate(Context, PasswordHashHash, 16)
      SHAUpdate(Context, NTResponse, 24)
      SHAUpdate(Context, Magic1, 39)
      SHAFinal(Context, Digest)

      ChallengeHash( PeerChallenge, AuthenticatorChallenge, UserName,
                     giving Challenge)

      SHAInit(Context)
      SHAUpdate(Context, Digest, 20)
      SHAUpdate(Context, Challenge, 8)
      SHAUpdate(Context, Magic2, 41)
      SHAFinal(Context, Digest)

      /*
       * Encode the value of 'Digest' as "S=" followed by
       * 40 ASCII hexadecimal digits and return it in
       * AuthenticatorResponse.
       * For example,
       *   "S=0123456789ABCDEF0123456789ABCDEF01234567"
       */

   }
    """
    Magic1="\x4D\x61\x67\x69\x63\x20\x73\x65\x72\x76\x65\x72\x20\x74\x6F\x20\x63\x6C\x69\x65\x6E\x74\x20\x73\x69\x67\x6E\x69\x6E\x67\x20\x63\x6F\x6E\x73\x74\x61\x6E\x74"
    Magic2="\x50\x61\x64\x20\x74\x6F\x20\x6D\x61\x6B\x65\x20\x69\x74\x20\x64\x6F\x20\x6D\x6F\x72\x65\x20\x74\x68\x61\x6E\x20\x6F\x6E\x65\x20\x69\x74\x65\x72\x61\x74\x69\x6F\x6E"

    password_hash=nt_password_hash(password,False)
    password_hash_hash=hash_nt_password_hash(password_hash)

    sha_hash=sha.new()
    sha_hash.update(password_hash_hash)
    sha_hash.update(nt_response)
    sha_hash.update(Magic1)
    digest=sha_hash.digest()
    
    challenge=challenge_hash(peer_challenge,authenticator_challenge,username)

    sha_hash=sha.new()
    sha_hash.update(digest)
    sha_hash.update(challenge)
    sha_hash.update(Magic2)
    digest=sha_hash.digest()
    
    return "\x01S="+convert_to_hex_string(digest)
    
def convert_to_hex_string(string):
    hex_str=""
    for c in string:
	hex_tmp=hex(ord(c))[2:]
	if len(hex_tmp)==1:
	    hex_tmp="0"+hex_tmp
	hex_str+=hex_tmp
    return hex_str.upper()
    
def hash_nt_password_hash(password_hash):
    """
   HashNtPasswordHash(
   IN  16-octet PasswordHash,
   OUT 16-octet PasswordHashHash )
   {
      /*
       * Use the MD4 algorithm [5] to irreversibly hash
       * PasswordHash into PasswordHashHash.
       */
   }
    """
    md4_context = md4.new()
    md4_context.update(password_hash)

    res = md4_context.digest()
    return res    

def lm_password_hash(password):
    """
   LmPasswordHash(
   IN  0-to-14-oem-char Password,
   OUT 16-octet         PasswordHash )
   {
      Set UcasePassword to the uppercased Password
      Zero pad UcasePassword to 14 characters

      DesHash( 1st 7-octets of UcasePassword,
               giving 1st 8-octets of PasswordHash )

      DesHash( 2nd 7-octets of UcasePassword,
               giving 2nd 8-octets of PasswordHash )
   }

    """
    ucase_password=password.upper()[:14]
    while len(ucase_password)<14:
	ucase_password+="\0"
    password_hash=des_hash(ucase_password[:7])
    password_hash+=des_hash(ucase_password[7:])
    return password_hash

def des_hash(clear):
    """
     DesHash(
   IN  7-octet Clear,
   OUT 8-octet Cypher )
   {
      /*
       * Make Cypher an irreversibly encrypted form of Clear by
       * encrypting known text using Clear as the secret key.
       * The known text consists of the string
       *
       *              KGS!@#$%
       */

      Set StdText to "KGS!@#$%"
      DesEncrypt( StdText, Clear, giving Cypher )
   }
    """
    des_obj=des.DES(clear)
    return des_obj.encrypt(r"KGS!@#$%")

