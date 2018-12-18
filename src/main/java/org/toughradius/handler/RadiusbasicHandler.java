package org.toughradius.handler;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.mina.core.buffer.IoBuffer;
import org.apache.mina.core.service.IoHandlerAdapter;
import org.apache.mina.core.session.IdleStatus;
import org.apache.mina.core.session.IoSession;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.tinyradius.attribute.RadiusAttribute;
import org.tinyradius.packet.AccessAccept;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusException;
import org.toughradius.component.*;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Nas;
import org.toughradius.entity.User;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.SocketAddress;
import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public abstract class RadiusbasicHandler extends IoHandlerAdapter {

    protected final static ScheduledExecutorService schedExecuter = Executors.newSingleThreadScheduledExecutor();

    @Autowired
    protected RadiusStat radiusStat;

    @Autowired
    protected RadiusConfig radiusConfig;

    @Autowired
    protected NasService brasService;

    @Autowired
    protected UserService subscribeService;

    @Autowired
    protected OnlineCache onlineCache;

    @Autowired
    protected UserCache subscribeCache;

    @Autowired
    protected OptionService configService;

    @Autowired
    protected RadiusParseFilter parseFilter;

    @Autowired
    protected ThreadPoolTaskExecutor taskExecutor;

    private final static Log logger = LogFactory.getLog(RadiusbasicHandler.class);

    /**
     * 查询设备信息
     * @param client
     * @param packet
     * @return
     * @throws RadiusException
     */
    public Nas getNas(InetSocketAddress client, RadiusPacket packet) throws RadiusException {
        String ip = client.getAddress().getHostAddress();
        RadiusAttribute nasid = packet.getAttribute(32);
        try {
            return brasService.findNas(ip,nasid.getAttributeValue());
        } catch (ServiceException e) {
            throw  new RadiusException(e.getMessage());
        }
    }

    /**
     * 查询用户信息
     * @param username
     * @return
     */
    public User getUser(String username) {
        return subscribeCache.findUser(username);
    }

    /**
     * 验证用户密码
     * @param user
     * @param accessRequest
     * @throws RadiusException
     */
    public void authUser(User user, AccessRequest accessRequest) throws RadiusException {
        String plaintext = user.getPassword();
        String ignorePwd = configService.getStringValue(OptionService.RADIUS_IGNORE_PASSWORD);

        if(!"enabled".equals(ignorePwd)){
            if (plaintext == null || !accessRequest.verifyPassword(plaintext)){
                throw new RadiusException("密码错误");
            }
        }
    }

    /**
     * 拷贝代理状态属性
     * @param request
     * @param answer
     */
    protected void copyProxyState(RadiusPacket request, RadiusPacket answer) {
        List proxyStateAttrs = request.getAttributes(33);
        for (Iterator i = proxyStateAttrs.iterator(); i.hasNext();) {
            RadiusAttribute proxyStateAttr = (RadiusAttribute) i.next();
            answer.addAttribute(proxyStateAttr);
        }
    }

    public RadiusPacket getAccountingResponse(AccountingRequest accountingRequest) throws RadiusException {
        RadiusPacket answer = new RadiusPacket(RadiusPacket.ACCOUNTING_RESPONSE, accountingRequest.getPacketIdentifier());
        copyProxyState(accountingRequest, answer);
        return answer;
    }

    public AccessAccept getAccessAccept(AccessRequest accessRequest) {
        AccessAccept answer = new AccessAccept(accessRequest.getPacketIdentifier());
        answer.addAttribute("Reply-Message","ok");
        copyProxyState(accessRequest, answer);
        return answer;
    }

    public RadiusPacket getAccessReject(AccessRequest accessRequest, String error) {
        RadiusPacket answer = new RadiusPacket(RadiusPacket.ACCESS_REJECT, accessRequest.getPacketIdentifier());
        answer.addAttribute("Reply-Message",error);
        copyProxyState(accessRequest, answer);
        return answer;
    }

    protected RadiusPacket makeRadiusPacket(byte[] data, String sharedSecret, int forceType) throws IOException, RadiusException {
        ByteArrayInputStream in = new ByteArrayInputStream(data);
        return RadiusPacket.decodeRequestPacket(in, sharedSecret, forceType);
    }

    /**
     * 发送延迟响应
     * @param delay
     * @param session
     * @param remoteAddress
     * @param secret
     * @param request
     * @param responses
     * @throws IOException
     */
    protected void sendDelayResponse(int delay, IoSession session, SocketAddress remoteAddress, String secret, RadiusPacket request, RadiusPacket responses) throws IOException {
        schedExecuter.schedule(()->{
            try {
                this.sendResponse(session,remoteAddress,secret,request,responses);
            } catch (IOException e) {
                logger.error("发送延迟响应失败",e);
            }
        },delay, TimeUnit.SECONDS);
    }

    /**
     * 发送正常响应
     * @param session
     * @param remoteAddress
     * @param secret
     * @param request
     * @param response
     * @throws IOException
     */
    protected void sendResponse(IoSession session, SocketAddress remoteAddress, String secret, RadiusPacket request, RadiusPacket response) throws IOException {
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        response.encodeResponsePacket(bos, secret, request);
        byte [] data = bos.toByteArray();
        IoBuffer outbuff = IoBuffer.wrap(data);
        radiusStat.incrRespBytes(data.length);
        session.write(outbuff,remoteAddress);
    }

    @Override
    public void exceptionCaught(IoSession session, Throwable cause)throws Exception {
        cause.printStackTrace();
        session.closeNow();
    }

    @Override
    public void sessionClosed(IoSession session) throws Exception {
    }

    @Override
    public void sessionCreated(IoSession session) throws Exception {
//        NioDatagramSessionConfig cfg = session.getConfig();
//        cfg.setReceiveBufferSize(64 * 1024 * 1024);
//        cfg.setReadBufferSize(2 * 1024 * 1024);
////        cfg.setKeepAlive(true);
//        cfg.setSoLinger(0);
    }

    @Override
    public void sessionIdle(IoSession session, IdleStatus status) throws Exception {
    }

    @Override
    public void sessionOpened(IoSession session) throws Exception {
    }


    /**
     * 重复报文判断
     * @param packet
     * @param address
     * @return
     */
    protected boolean isPacketDuplicate(RadiusPacket packet, InetSocketAddress address) {
        long now = System.currentTimeMillis();
        long intervalStart = now - getDuplicateInterval();
        byte[] authenticator = packet.getAuthenticator();
        for (Iterator i = receivedPackets.iterator(); i.hasNext();) {
            ReceivedPacket p = (ReceivedPacket) i.next();
            if (p.receiveTime < intervalStart) {
                // packet is older than duplicate interval
                i.remove();
            }
            else {
                if (p.address.equals(address) && p.packetIdentifier == packet.getPacketIdentifier()) {
                    if (authenticator != null && p.authenticator != null) {
                        // packet is duplicate if stored authenticator is equal
                        // to the packet authenticator
                        return Arrays.equals(p.authenticator, authenticator);
                    }
                    // should not happen, packet is duplicate
                    return true;
                }
            }
        }

        // add packet to receive list
        ReceivedPacket rp = new ReceivedPacket();
        rp.address = address;
        rp.packetIdentifier = packet.getPacketIdentifier();
        rp.receiveTime = now;
        rp.authenticator = authenticator;
        receivedPackets.add(rp);


        return false;
    }

    public long getDuplicateInterval() {
        return duplicateInterval;
    }
    private Queue receivedPackets = new ConcurrentLinkedQueue();
    private long duplicateInterval = 30000; // 30 s

    class ReceivedPacket {

        /**
         * The identifier of the packet.
         */
        public int packetIdentifier;

        /**
         * The time the packet was received.
         */
        public long receiveTime;

        /**
         * The address of the host who sent the packet.
         */
        public InetSocketAddress address;

        /**
         * Authenticator of the received packet.
         */
        public byte[] authenticator;

    }
}
