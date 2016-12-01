try:
   import cPickle as pickle
except:
   import pickle
import time
import hmac
import uuid
import hashlib
import base64
from sqlalchemy.sql import text as _sql

class SessionData(dict):
    def __init__(self, session_id, hmac_key):
        self.session_id = session_id
        self.hmac_key = hmac_key
    
class Session(SessionData):
    def __init__(self, session_manager, request_handler):
        self.session_manager = session_manager
        self.request_handler = request_handler
        
        try:
            current_session = session_manager.get(request_handler)
        except InvalidSessionException:
            current_session = session_manager.get()
            
        for key, data in current_session.iteritems():
            self[key] = data
        self.session_id = current_session.session_id
        self.hmac_key = current_session.hmac_key
    
    def save(self):
        self.session_manager.set(self.request_handler, self)

    def clear(self):
        self.session_manager.clear(self.request_handler, self)


class SessionManager(object):
    def __init__(self, secret, dbengine, session_timeout):
        self.secret = secret
        self.dbengine = dbengine
        self.session_timeout = session_timeout

    def encode_data(self,data):
        return base64.b64encode(pickle.dumps(data, pickle.HIGHEST_PROTOCOL))

    def decode_data(self,raw_data):
        return pickle.loads(base64.b64decode(raw_data))

    def _raw_get(self, key, **kwargs):
        raw_data = None
        with self.dbengine.begin() as conn:
            try:
                cur = conn.execute(_sql("select _value, _time from system_session where _key = :key "),key=key)
                session =  cur.fetchone()
                if session:
                    _time = int(session['_time'])
                    if (time.time() - _time) > self.session_timeout:
                        conn.execute(_sql("delete from system_session where _key = :key "),key=key)
                    else:
                        raw_data = session['_value']
            except:
                import traceback
                traceback.print_exc()
        return raw_data

    def _raw_set(self, key, value, timeout,**kwargs):
        with self.dbengine.begin() as conn:
            _time = int(time.time()) + int(timeout)
            try:
                conn.execute(_sql("insert into system_session values (:key, :value, :time) "),
                    key=key,value=value,time=_time)
            except:
                conn.execute(_sql("""update system_session 
                                    set _value=:value, _time=:time
                                    where _key=:key"""),
                                    key=key,value=value,time=_time)

    def _delete(self, key):
        with self.dbengine.begin() as conn:
            try:
                conn.execute(_sql("delete from system_session where _key = :key "),key=key)
            except:
                import traceback
                traceback.print_exc()

    def _raw_replace(self, key, value, timeout,**kwargs):
        with self.dbengine.begin() as conn:
            _time = int(time.time()) + int(timeout)
            try:
                conn.execute(_sql("""update system_session 
                                    set _value=:value, _time=:time
                                    where _key=:key"""),
                                    key=key,value=value,time=_time)
            except:
                import traceback
                traceback.print_exc()

        
    def _fetch(self, session_id):
        try:
            session_data = raw_data = self._raw_get(session_id)
            if raw_data != None:
                self._raw_replace(session_id, raw_data, self.session_timeout)
                session_data = self.decode_data(raw_data)
            if type(session_data) == type({}):
                return session_data
            else:
                return {}
        except IOError:
            return {}
        
    def get(self, request_handler = None):
        
        if (request_handler == None):
            session_id = None
            hmac_key = None
        else:
            session_id = request_handler.get_secure_cookie("session_id")
            hmac_key = request_handler.get_secure_cookie("verification")
        
        if session_id == None:
            session_exists = False
            session_id = self._generate_id()
            hmac_key = self._generate_hmac(session_id)
        else:
            session_exists = True
            
        check_hmac = self._generate_hmac(session_id)
        if hmac_key != check_hmac:
            raise InvalidSessionException()
        
        session = SessionData(session_id, hmac_key)
        
        if session_exists:
            session_data = self._fetch(session_id)
            for key, data in session_data.iteritems():
                session[key] = data
                
        return session
    
    def set(self, request_handler, session):
        request_handler.set_secure_cookie("session_id", session.session_id)
        request_handler.set_secure_cookie("verification", session.hmac_key)
        session_data = self.encode_data(dict(session.items()))
        self._raw_set(session.session_id, session_data, self.session_timeout)   

    def clear(self, request_handler, session):
        request_handler.clear_all_cookies()  
        self._delete(session.session_id)
        
    def _generate_id(self):
        new_id = hashlib.sha256(self.secret + str(uuid.uuid4()))
        return new_id.hexdigest()
    
    def _generate_hmac(self, session_id):
        return hmac.new(session_id, self.secret, hashlib.sha256).hexdigest()

class InvalidSessionException(Exception):
    pass