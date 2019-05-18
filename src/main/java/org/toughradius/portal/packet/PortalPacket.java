package org.toughradius.portal.packet;

import org.apache.mina.core.buffer.IoBuffer;
import org.toughradius.common.ValidateUtil;
import org.toughradius.portal.PortalException;
import org.toughradius.portal.utils.PortalUtils;

import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.*;
import java.util.concurrent.atomic.AtomicInteger;

/**
 REQ_CHALLENGE	0x01	Portal server--> BAS	表示此报文是Portal Server向BAS发送的 Challenge请求报文	必须
 ACK_CHALLENGE	0x02	BAS --> Portal server	表示此报文是BAS对Portal Server请求Challenge报文的响应报文	必须
 REQ_AUTH	0x03	Portal server --> BAS	表示此报文是Portal Server向BAS发送的请求认证报文	必须
 ACK_AUTH	0x04	BAS --> Portal server	表示此报文是BAS对Portal Server请求认证报文的响应报文	必须
 REQ_LOGOUT	0x05	Portal server --> BAS	表示此报文是Portal  Server向BAS发送的下线请求报文	必须
 ACK_LOGOUT	0x06	BAS --> Portal server	表示此报文是BAS对Portal Server下线请求的响应报文	必须
 AFF_ACK_AUTH	0x07	Portal server --> BAS	表示此报文是Portal Server收到认证成功响应报文后向BAS发送的确认报文	建议
 NTF_LOGOUT	0x08	BAS --> Portal server	表示此报文是BAS发送给Portal Server，用户被强制下线的通知报文	必须
 REQ_INFO	0x09	Portal server --> BAS	信息询问报文	必须
 ACK_INFO 	0x0a	BAS --> Portal server	信息询问的应答报文	必须
 NTF_USERDISCOVER	0x0b	Portal server --> BAS	Portal Server向BAS发送的发现新用户要求上线的通知报文	建议
 NTF_USERIPCHANGE 	0x0c	BAS --> Portal server	BAS向Portal Server发送的通知更改某个用户IP地址的通知报文	必须
 AFF_NTF_USERIPCHAN	0x0d	Portal server --> BAS	PortalServer通知BAS对用户表项的IP切换已成功	必须
 ACK_NTF_LOGOUT	0x0e	Portal server --> BAS	PortalServer通知BAS用户强制下线成功，BAS通过NTF_LOGOUT报文通知Portal Server用户下线后，Portal Server回应BAS设备用户下线完成的回应报文。如果Portal Server收到了BAS的用户下线请求，必须回应ACK_NTF_LOGOUT，以通知BAS服务器，无论用户是否在线。同时，Portal Server必须确保用户下线处理成功。	必须
 NTF_HEARTBEAT	0x0f	Portal server --> BAS	逃生心跳报文，PortalServer周期性的向BAS发送该报文，以表明PortalServer可以正常提供服务。BAS如果连续多次没有接收到该报文，说明PortalServer已经停止服务，BAS即切换为逃生状态，此时不再强制用户认证，允许用户的报文直接通过。该报文没有回应报文。	必须
 NTF_USER_HEARTBEAT	0x10	Portal server --> BAS	用户心跳报文，PortalServer周期性的向BAS发送该报文，以表明该用户仍然在线，BAS如果连续多次没有接收到含有该用户IP的报文，说明该用户已经断线，BAS将向RADIUS服务器发送下线报文，将用户下线。用户心跳报文中包含了多个用户的IP地址。	必须
 ACK_NTF_USER_HEARTBEAT	0x11	BAS --> Portal server	用户心跳回应报文，BAS接收到PortalServer的用户心跳报文后，会遍历这些用户IP地址，并将已经下线的用户IP地址放入回应报文中。PortalServer收到回应报文后，将用户下线。如果用户心跳报文中的所有用户都在线，则BAS将不发送回应报文。	必须
 NTF_CHALLENGE	0x12	BAS --> Portal server	表示此报文是 BAS 向Portal Server 发送的Challenge请求报文，主要适用于EAP_TLS认证。	建议
 NTF_USER_NOTIFY	0x13	BAS --> Portal server	用户消息通知报文。在Pap/Chap认证方式下，计费回应报文中Radius服务器需要向用户下发一些消息，例如帐号余额等信息。	建议
 AFF_NTF_USER_NOTIFY	0x14	Portal server --> BAS	PortalServer通知BAS消息已收到	建议
 */
public  class PortalPacket {

    /**
     * 消息类型定义
     */
    public final static int REQ_CHALLENGE = 0x01;
    public final static int ACK_CHALLENGE = 0x02;
    public final static int REQ_AUTH = 0x03;
    public final static int ACK_AUTH = 0x04;
    public final static int REQ_LOGOUT = 0x05;
    public final static int ACK_LOGOUT = 0x06;
    public final static int AFF_ACK_AUTH = 0x07;
    public final static int NTF_LOGOUT = 0x08;
    public final static int REQ_INFO = 0x09;
    public final static int ACK_INFO = 0x0a;
    public final static int NTF_USERDISCOVER = 0x0b;
    public final static int NTF_USERIPCHANGE = 0x0c;
    public final static int AFF_NTF_USERIPCHAN = 0x0d;
    public final static int ACK_NTF_LOGOUT = 0x0e;
    public final static int NTF_HEARTBEAT = 0x0f;
    public final static int NTF_USER_HEARTBEAT = 0x10;
    public final static int ACK_NTF_USER_HEARTBEAT = 0x11;
    public final static int NTF_CHALLENGE = 0x12;
    public final static int NTF_USER_NOTIFY = 0x13;
    public final static int AFF_NTF_USER_NOTIFY = 0x14;

    /**
     * CHAP/PAP
     */
    public final static int AUTH_CHAP = 0x00;
    public final static int AUTH_PAP = 0x01;

    /**
     * ATTRIBUTE TYPE
     */
    public final static int ATTRIBUTE_MAC_TYPE = 0xff;
    public final static int ATTRIBUTE_USERNAME_TYPE = 0x01;
    public final static int ATTRIBUTE_PASSWORD_TYPE = 0x02;
    public final static int ATTRIBUTE_CHALLENGE_TYPE = 0x03;
    public final static int ATTRIBUTE_CHAP_PWD_TYPE = 0x04;
    public final static int ATTRIBUTE_TEXT_INFO_TYPE = 0x05;
    public final static int ATTRIBUTE_UP_LINK_TYPE = 0x06;
    public final static int ATTRIBUTE_DOWN_LINK_TYPE = 0x07;
    public final static int ATTRIBUTE_PORT_TYPE = 0x08;
    public final static int ATTRIBUTE_BASIP_TYPE = 0x0a;

    /**
     * PACKET VER
     */
    public final static int CMCCV1_TYPE = 0x01;
    public final static int CMCCV2_TYPE = 0x01;
    public final static int HUAWEIV1_TYPE = 0x01;
    public final static int HUAWEIV2_TYPE = 0x02;

    public static final int MAX_PACKET_LENGTH = 1024;
    public static final Map<Integer,String> ACK_CHALLENGE_ERRORS = new HashMap<Integer,String>();
    public static final Map<Integer,String> ACK_AUTH_ERRORS = new HashMap<Integer,String>();
    public static final Map<Integer,String> ACK_LOGOUT_ERRORS = new HashMap<Integer,String>();
    public static final Map<Integer,String> ACK_INFO_ERRORS = new HashMap<Integer,String>();

    static{
        //ACK_CHALLENGE_ERRORS
        ACK_CHALLENGE_ERRORS.put(0,"请求Challenge成功");
        ACK_CHALLENGE_ERRORS.put(1,"请求Challenge被拒绝");
        ACK_CHALLENGE_ERRORS.put(2,"链接已经建立");
        ACK_CHALLENGE_ERRORS.put(3,"有一个用户正在认证过程中，请稍后再试");
        ACK_CHALLENGE_ERRORS.put(4, "请求Challenge失败");
        //ACK_AUTH_ERRORS
        ACK_AUTH_ERRORS.put(0,"用户认证请求成功");
        ACK_AUTH_ERRORS.put(1,"用户认证请求被拒绝");
        ACK_AUTH_ERRORS.put(2,"链接已经建立");
        ACK_AUTH_ERRORS.put(3,"有一个用户正在认证过程中，请稍后再试");
        ACK_AUTH_ERRORS.put(4,"用户认证失败");
        //ACK_LOGOUT_ERRORS
        ACK_LOGOUT_ERRORS.put(0,"用户下线成功");
        ACK_LOGOUT_ERRORS.put(1,"用户下线被拒绝");
        ACK_LOGOUT_ERRORS.put(2,"用户下线失败");
        ACK_LOGOUT_ERRORS.put(3,"用户已经离线");
        //ACK_INFO_ERRORS
        ACK_INFO_ERRORS.put(0,"SUCCESS");
        ACK_INFO_ERRORS.put(1,"功能不被支持");
        ACK_INFO_ERRORS.put(2,"消息处理失败");
    }

    private int ver = 0x01;
    private int type;
    private int isChap = AUTH_PAP;
    private int rsv = 0;
    private short serialNo = 0;
    private short reqId = 0;
    private String userIp;
    private short userPort = 0;
    private int errCode = 0;
    private int attrNum = 0;
    private byte[] authenticator;

    private String secret;

    private List<PortalAttribute> attributes = new ArrayList<PortalAttribute>();

    private static AtomicInteger nextSerialNo = new AtomicInteger(0);
    private static AtomicInteger nextReqId = new AtomicInteger(0);
    private static SecureRandom random = new SecureRandom();
    private MessageDigest md5Digest = null;


    public static short getNextSerialNo() {
        int val = nextSerialNo.incrementAndGet();
        if (val >= Short.MAX_VALUE){
            nextSerialNo.set(0);
            val = 0;
        }
        return (short) val;
    }
    public static short getNextReqId() {
        int val = nextReqId.incrementAndGet();
        if (val >= Short.MAX_VALUE){
            nextReqId.set(0);
            val = 0;
        }
        return (short) val;
    }

    public String getErrMessage(){
        if(getType() == 1 || getType() == 3 || getType()==7){
            return "";
        }

        switch (getType()){
            case ACK_CHALLENGE:return ACK_CHALLENGE_ERRORS.get(getErrCode());
            case ACK_AUTH:return ACK_AUTH_ERRORS.get(getErrCode());
            case ACK_LOGOUT:return ACK_LOGOUT_ERRORS.get(getErrCode());
            case ACK_INFO:return ACK_INFO_ERRORS.get(getErrCode());
            default:
                return "";
        }

    }

    public static int getVerbyName(String name){
        switch (name){
            case "cmccv1":return CMCCV1_TYPE;
            case "cmccv2":return CMCCV2_TYPE;
            case "huaweiv1":return HUAWEIV1_TYPE;
            case "huaweuiv2":return HUAWEIV2_TYPE;
            default:return CMCCV1_TYPE;
        }
    }

    public String getPacketTypeName(){
        switch (getType()){
            case REQ_CHALLENGE:return "REQ_CHALLENGE";
            case ACK_CHALLENGE:return "ACK_CHALLENGE";
            case REQ_AUTH:return "REQ_AUTH";
            case ACK_AUTH:return "ACK_AUTH";
            case REQ_LOGOUT:return "REQ_LOGOUT";
            case ACK_LOGOUT:return "ACK_LOGOUT";
            case AFF_ACK_AUTH:return "AFF_ACK_AUTH";
            case NTF_LOGOUT:return "NTF_LOGOUT";
            case REQ_INFO:return "REQ_INFO";
            case ACK_INFO:return "ACK_INFO";
            case NTF_USERDISCOVER:return "NTF_USERDISCOVER";
            case NTF_USERIPCHANGE:return "NTF_USERIPCHANGE";
            case AFF_NTF_USERIPCHAN:return "AFF_NTF_USERIPCHAN";
            case ACK_NTF_LOGOUT:return "ACK_NTF_LOGOUT";
            case NTF_HEARTBEAT:return "NTF_HEARTBEAT";
            case NTF_USER_HEARTBEAT:return "NTF_USER_HEARTBEAT";
            case ACK_NTF_USER_HEARTBEAT:return "ACK_NTF_USER_HEARTBEAT";
            case NTF_CHALLENGE:return "NTF_CHALLENGE";
            case NTF_USER_NOTIFY:return "NTF_USER_NOTIFY";
            case AFF_NTF_USER_NOTIFY:return "AFF_NTF_USER_NOTIFY";
            default:
                return "Unknown (" + getType() + ")";
        }
    }

    public PortalPacket() {
    }

    public PortalPacket(byte [] src) throws PortalException {
        this.decodePacket(IoBuffer.wrap(src));
    }

    public PortalPacket(int ver,int type, String userIp, short serialNo, short reqId, String secret, int isChap) {
        setVer(ver);
        setType(type);
        setUserIp(userIp);
        setSerialNo(serialNo);
        setReqId(reqId);
        setSecret(secret);
        setIsChap(isChap);
    }

    /**
     * 编码报文
     * @return
     * @throws PortalException
     */
    public IoBuffer encodePacket() throws PortalException {
        IoBuffer buffer = IoBuffer.allocate(16);
        buffer.setAutoExpand(true);
        buffer.put((byte)getVer());
        buffer.put((byte)getType());
        buffer.put((byte)getIsChap());
        buffer.put((byte)getRsv());
        buffer.putShort(getSerialNo());
        buffer.putShort(getReqId());
        buffer.put(PortalUtils.encodeIpV4(getUserIp()));
        buffer.putShort(getUserPort());
        buffer.put((byte)getErrCode());
        buffer.put((byte)getAttrNum());
        if(getVer() == HUAWEIV2_TYPE){
            byte[] auth = getAuthenticator();
            if(auth==null){
                throw new PortalException("Request authenticator is empty");
            }
            buffer.put(getAuthenticator());
        }
        for(PortalAttribute attr : getAttributes()){
            buffer.put(attr.encodeAttribute());
        }
        buffer.flip();
        return buffer;
    }

    /**
     * 报文解码
     * @param buff
     * @throws PortalException
     */
    public void decodePacket(IoBuffer buff) throws PortalException {
        buff.rewind();
        if(buff.remaining()>MAX_PACKET_LENGTH){
            throw new PortalException("Packet size is too large");
        }
        byte ver = buff.get();
        if(ver!=HUAWEIV1_TYPE&&ver!=HUAWEIV2_TYPE){
            throw new PortalException("Packet ver error");
        }
        setVer(ver);
        setType(buff.get());
        byte ischap = buff.get();
        if(ischap!=AUTH_CHAP&&ischap!=AUTH_PAP){
            throw new PortalException("Packet chap/pap error");
        }
        setIsChap(ischap);
        setRsv(buff.get());
        setSerialNo(buff.getShort());
        setReqId(buff.getShort());
        byte [] userIpdata = new byte[4];
        buff.get(userIpdata);
        setUserIp(PortalUtils.decodeIpv4(userIpdata));
        setUserPort(buff.getShort());
        setErrCode(buff.get());
        setAttrNum(buff.get());
        if(getVer()==HUAWEIV2_TYPE){
            byte[] auth = new byte[16];
            buff.get(auth);
            setAuthenticator(auth);
        }
        for(int i=0;i<getAttrNum();i++){
            PortalAttribute attr = new PortalAttribute();
            attr.setAttributeType(buff.get());
            int len = (int) buff.get();
            if(len==2)
                continue;
            if (len==0){
                continue;
            }
            byte[] attrdata = new byte[len-2];
            buff.get(attrdata);
            attr.setAttributeData(attrdata);
        }
    }



    /**
     * 创建请求验证字
     * @param sharedSecret
     * @return
     */
    protected byte[] createRequestAuthenticator(String sharedSecret) {
        byte[] secretBytes = PortalUtils.encodeString(sharedSecret);
        byte[] randomBytes = new byte[16];
        MessageDigest md5 = getMd5Digest();
        md5.reset();
        md5.update((byte)getVer());
        md5.update((byte)getType());
        md5.update((byte)getIsChap());
        md5.update((byte)getRsv());
        md5.update(PortalUtils.encodeShort(getSerialNo()));
        md5.update(PortalUtils.encodeShort(getReqId()));
        md5.update(PortalUtils.encodeIpV4(getUserIp()));
        md5.update(PortalUtils.encodeShort(getUserPort()));
        md5.update((byte)getErrCode());
        md5.update((byte)attributes.size());
        md5.update(randomBytes);
        for(PortalAttribute attr : attributes){
            md5.update(attr.encodeAttribute());
        }
        md5.update(PortalUtils.encodeString(sharedSecret));
        return md5.digest();
    }

    protected void updateRequestAuthenticator(String secret) {
        if(getVer() == HUAWEIV2_TYPE){
            setAuthenticator(createRequestAuthenticator(secret));
        }
    }

    public void updateResponseAuthenticator(String secret) {
        if(getVer() == HUAWEIV2_TYPE){
            setAuthenticator(createRequestAuthenticator(secret));
        }
    }

    /**
     * 创建响应验证字
     * @param sharedSecret
     * @param requestAuthenticator
     * @return
     */
    protected byte[] createResponseAuthenticator(String sharedSecret, byte[] requestAuthenticator) {
        MessageDigest md5 = getMd5Digest();
        md5.reset();
        md5.update((byte)getVer());
        md5.update((byte)getType());
        md5.update((byte)getIsChap());
        md5.update((byte)getRsv());
        md5.update(PortalUtils.encodeShort(getSerialNo()));
        md5.update(PortalUtils.encodeShort(getReqId()));
        md5.update(PortalUtils.encodeIpV4(getUserIp()));
        md5.update(PortalUtils.encodeShort(getUserPort()));
        md5.update((byte)getErrCode());
        md5.update((byte)attributes.size());
        md5.update(requestAuthenticator);
        for(PortalAttribute attr : attributes){
            md5.update(attr.encodeAttribute());
        }
        md5.update(PortalUtils.encodeString(sharedSecret));
        return md5.digest();
    }

    /**
     * 校验响应验证字
     * @param sharedSecret
     * @param requestAuthenticator
     */
    public void checkResponseAuthenticator(String sharedSecret,  byte[] requestAuthenticator)throws PortalException {
        if(requestAuthenticator==null)
            return;
        byte[] authenticator = createResponseAuthenticator(sharedSecret, requestAuthenticator);
        byte[] receivedAuth = getAuthenticator();
        for (int i = 0; i < 16; i++)
            if (authenticator[i] != receivedAuth[i])
                throw new PortalException("response authenticator invalid");
    }


    ///////////////////////////////////////////////////////////////////////////////////////

    /**
     * 创建新报文
     * @param ver
     * @param type
     * @param userIp
     * @param serialNo
     * @param reqId
     * @param secret
     * @param isChap
     * @return
     */
    public static PortalPacket createMessage(int ver, int type, String userIp, short serialNo, short reqId, String secret, int isChap){
        PortalPacket message =  new PortalPacket(ver,type,userIp,serialNo,reqId,secret,isChap);
        message.updateRequestAuthenticator(secret);
        return message;
    }


    /**
     * 向BAS发送的 Challenge请求报文 (必须)
     * @param ver
     * @param userIp
     * @param secret
     * @param mac
     * @return
     */
    public static PortalPacket createReqChallenge(int ver, String userIp, String secret, String mac){
        PortalPacket message = createMessage(ver, REQ_CHALLENGE,userIp,getNextSerialNo(),(short)0,secret, AUTH_CHAP);
        if(ValidateUtil.isMacAddress(mac)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_MAC_TYPE, PortalUtils.encodeMacAddr(mac)));
        }
        return message;
    }

    /**
     * Portal Server向BAS发送的请求认证报文 (必须)
     * @param userIp
     * @param username
     * @param password
     * @param reqId
     * @param challenge
     * @param secret
     * @param basIp
     * @param isChap
     * @param mac
     * @return
     */
    public static PortalPacket createReqAuth(int ver, String userIp, String username, String password, short reqId,
                                             byte[] challenge, String secret, String  basIp, int isChap, String mac){
        PortalPacket message = createMessage(ver,REQ_AUTH,userIp,getNextSerialNo(),reqId,secret, isChap);
        message.addAttribute(new PortalAttribute(ATTRIBUTE_USERNAME_TYPE, PortalUtils.encodeString(username)));

        if(isChap == AUTH_CHAP){
            byte[] userPassword = PortalUtils.chapEncryption(password, reqId, challenge);
            message.addAttribute(new PortalAttribute(ATTRIBUTE_CHAP_PWD_TYPE,userPassword));
        }else{
            message.addAttribute(new PortalAttribute(ATTRIBUTE_PASSWORD_TYPE, PortalUtils.encodeString(password)));
        }

        if(ValidateUtil.isMacAddress(mac)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_MAC_TYPE, PortalUtils.encodeMacAddr(mac)));
        }

        if(ValidateUtil.isIP(basIp)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_BASIP_TYPE, PortalUtils.encodeIpV4(basIp)));
        }
        return message;
    }

    /**
     * Portal  Server向BAS发送的下线请求报文 (必须)
     * @param userIp
     * @param secret
     * @param serialNo
     * @param isChap
     * @param mac
     * @return
     */
    public static PortalPacket createReqLogout(int ver, String userIp, String secret, String basIp, short serialNo, int isChap, String mac){
        short _serialNo = serialNo==-1?getNextSerialNo():serialNo;
        PortalPacket message = createMessage(ver, REQ_LOGOUT,userIp,_serialNo,(short)0,secret, isChap);

        if(ValidateUtil.isMacAddress(mac)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_MAC_TYPE, PortalUtils.encodeMacAddr(mac)));
        }

        if(ValidateUtil.isIP(basIp)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_BASIP_TYPE, PortalUtils.encodeIpV4(basIp)));
        }
        return message;
    }

    /**
     * 收到认证成功响应报文后向BAS发送的确认报文, （可选）
     * @param ver
     * @param userIp
     * @param secret
     * @param basIp
     * @param serialNo
     * @param reqId
     * @param isChap
     * @param mac
     * @return
     */
    public static PortalPacket createAffAckAuth(int ver, String userIp, String secret, String basIp, short serialNo, short reqId, int isChap, String mac){
        short _serialNo = serialNo==-1?getNextSerialNo():serialNo;
        PortalPacket message = createMessage(ver,AFF_ACK_AUTH,userIp,_serialNo,reqId,secret, isChap);

        if(ValidateUtil.isMacAddress(mac)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_MAC_TYPE, PortalUtils.encodeMacAddr(mac)));
        }

        if(ValidateUtil.isIP(basIp)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_BASIP_TYPE, PortalUtils.encodeIpV4(basIp)));
        }
        return message;
    }

    /**
     * 信息询问报文 (必须)
     * @param ver
     * @param userIp
     * @param secret
     * @param basIp
     * @param serialNo
     * @param isChap
     * @param mac
     * @return
     */
    public static PortalPacket createReqInfo(int ver, String userIp, String secret, String basIp, short serialNo, int isChap, String mac){
        short _serialNo = serialNo==-1?getNextSerialNo():serialNo;
        PortalPacket message = createMessage(ver, REQ_INFO,userIp,_serialNo,(short)0,secret, isChap);

        if(ValidateUtil.isMacAddress(mac)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_MAC_TYPE, PortalUtils.encodeMacAddr(mac)));
        }

        if(ValidateUtil.isIP(basIp)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_BASIP_TYPE, PortalUtils.encodeIpV4(basIp)));
        }
        return message;
    }

    /**
     * 逃生心跳报文，PortalServer周期性的向BAS发送该报文，以表明PortalServer可以正常提供服务。
     * BAS如果连续多次没有接收到该报文，说明PortalServer已经停止服务，BAS即切换为逃生状态，此时不再强制用户认证，允许用户的报文直接通过。该报文没有回应报文
     * @param ver
     * @param secret
     * @param basIp
     * @param isChap
     * @param mac
     * @return
     */
    public static PortalPacket createNtfHeart(int ver, String secret, String basIp, int isChap, String mac){
        PortalPacket message = createMessage(ver, NTF_HEARTBEAT,"0.0.0.0",getNextSerialNo(),(short)0,secret, isChap);

        if(ValidateUtil.isMacAddress(mac)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_MAC_TYPE, PortalUtils.encodeMacAddr(mac)));
        }

        if(ValidateUtil.isIP(basIp)){
            message.addAttribute(new PortalAttribute(ATTRIBUTE_BASIP_TYPE, PortalUtils.encodeIpV4(basIp)));
        }
        return message;
    }


    //////////////////////////////////////////////////////////////////////////////////////
    public String getUsername(){
        for(Iterator<PortalAttribute> it = attributes.iterator();it.hasNext();){
            PortalAttribute attr = it.next();
            if(attr.getAttributeType() == ATTRIBUTE_USERNAME_TYPE){
                return attr.getAttributeAsStr();
            }
        }
        return null;
    }

    public String getPassword(){
        for(Iterator<PortalAttribute> it = attributes.iterator();it.hasNext();){
            PortalAttribute attr = it.next();
            if(attr.getAttributeType() == ATTRIBUTE_PASSWORD_TYPE){
                return attr.getAttributeAsStr();
            }
        }
        return null;
    }

    public byte[] getChallenge(){
        for(Iterator<PortalAttribute> it = attributes.iterator();it.hasNext();){
            PortalAttribute attr = it.next();
            if(attr.getAttributeType() == ATTRIBUTE_CHALLENGE_TYPE){
                return attr.getAttributeData();
            }
        }
        return null;
    }

    public byte[] getChapPassword(){
        for(Iterator<PortalAttribute> it = attributes.iterator();it.hasNext();){
            PortalAttribute attr = it.next();
            if(attr.getAttributeType() == ATTRIBUTE_CHAP_PWD_TYPE){
                return attr.getAttributeData();
            }
        }
        return null;
    }

    public String getTextInfo(){
        for(Iterator<PortalAttribute> it = attributes.iterator();it.hasNext();){
            PortalAttribute attr = it.next();
            if(attr.getAttributeType() == ATTRIBUTE_CHAP_PWD_TYPE){
                return attr.getAttributeAsStr();
            }
        }
        return null;
    }

    public String getBasIp(){
        for(Iterator<PortalAttribute> it = attributes.iterator();it.hasNext();){
            PortalAttribute attr = it.next();
            if(attr.getAttributeType() == ATTRIBUTE_BASIP_TYPE){
                return PortalUtils.decodeIpv4(attr.getAttributeData());
            }
        }
        return null;
    }



    public String toString() {
        StringBuffer s = new StringBuffer();
        s.append(String.format("%s -> ", getPacketTypeName()));
        s.append(String.format("Ver=%s,", getVer()));
        s.append(String.format("Type=%s,", getType()));
        s.append(String.format("Chap/Pap=%s,", getIsChap()));
        s.append(String.format("SerialNo=%s,", getSerialNo()));
        s.append(String.format("ReqId=%s,", getReqId()));
        s.append(String.format("UserIp=%s,", getUserIp()));
        s.append(String.format("UserPort=%s,", getUserPort()));
        s.append(String.format("ErrCode=%s,", getErrCode()));
        s.append(String.format("AttrNum=%s,", attributes.size()));
        s.append("\nAttributes::::");
        for (Iterator i = attributes.iterator(); i.hasNext();) {
            PortalAttribute attr = (PortalAttribute) i.next();
            s.append("\n");
            s.append(String.format("\t%s", attr.toString()));
        }
        return s.toString();
    }

    public byte[] getAuthenticator() {
        return authenticator;
    }

    public void setAuthenticator(byte[] authenticator) {
        this.authenticator = authenticator;
    }

    public List<PortalAttribute> getAttributes() {
        return attributes;
    }

    public void addAttribute(PortalAttribute attribute){
        attributes.add(attribute);
    }

    public int getVer() {
        return ver;
    }

    public void setVer(int ver) {
        this.ver = ver;
    }

    public int getType() {
        return type;
    }

    public void setType(int type) {
        this.type = type;
    }

    public int getIsChap() {
        return isChap;
    }

    public void setIsChap(int isChap) {
        this.isChap = isChap;
    }

    public int getRsv() {
        return rsv;
    }

    public void setRsv(int rsv) {
        this.rsv = rsv;
    }

    public short getSerialNo() {
        return serialNo;
    }

    public void setSerialNo(short serialNo) {
        this.serialNo = serialNo;
    }

    public short getReqId() {
        return reqId;
    }

    public void setReqId(short reqId) {
        this.reqId = reqId;
    }

    public String getUserIp() {
        return userIp;
    }

    public void setUserIp(String userIp) {
        this.userIp = userIp;
    }

    public short getUserPort() {
        return userPort;
    }

    public void setUserPort(short userPort) {
        this.userPort = userPort;
    }

    public int getErrCode() {
        return errCode;
    }


    public int getAttrNum() {
        return attributes.size();
    }

    public void setAttrNum(int attrNum) {
        this.attrNum = attrNum;
    }

    public String getSecret() {
        return secret;
    }

    public void setSecret(String secret) {
        this.secret = secret;
    }

    public void setErrCode(int errCode) {
        this.errCode = errCode;
    }

    protected MessageDigest getMd5Digest() {
        if (md5Digest == null)
            try {
                md5Digest = MessageDigest.getInstance("MD5");
            }
            catch (NoSuchAlgorithmException nsae) {
                throw new RuntimeException("md5 digest not available", nsae);
            }
        return md5Digest;
    }
}
