package org.toughradius.config;
import org.toughradius.common.ValidateCache;
import org.toughradius.handler.RadiusAcctHandler;
import org.toughradius.handler.RadiusBasicHandler;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.mina.core.filterchain.DefaultIoFilterChainBuilder;
import org.apache.mina.core.filterchain.IoFilter;
import org.apache.mina.filter.executor.ExecutorFilter;
import org.apache.mina.transport.socket.DatagramSessionConfig;
import org.apache.mina.transport.socket.nio.NioDatagramAcceptor;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@Configuration
@ConfigurationProperties(prefix = "org.toughradius")
public class RadiusConfig {

    private Log logger = LogFactory.getLog(RadiusConfig.class);

    private int authport;
    private int acctport;
    private int eventport;
    private int trace;
    private int interimUpdate;
    private int maxSessionTimeout;
    private String ticketDir;
    private String statDir;
    private boolean running;
    private boolean isBillInput;
    private boolean isBillBackFlow;
    private boolean allowNegative;
    private int rejectdelay;
    private int rejectdelayTimes;
    private int rejectdelayEnabled;
    private int ticketExpireDays;
    private int authPool;
    private int acctPool;
    private String statfile;

    /**
     * 5秒内认证拒绝超过 rejectdelayTimes 次数，将触发延迟拒绝
     * @return
     */
    @Bean
    public ValidateCache authValidate(){
        return new ValidateCache(5000,rejectdelayTimes);
    }

    /**
     * Radius 认证服务配置
     * @param radiusAuthHandler
     * @return
     * @throws IOException
     */
    @Bean( destroyMethod = "unbind")
    public NioDatagramAcceptor nioAuthAcceptor(RadiusBasicHandler radiusAuthHandler) throws IOException {
        if(!running){
            logger.info("====== RadiusAuthServer not running =======");
            return null;
        }
        NioDatagramAcceptor nioAuthAcceptor = new NioDatagramAcceptor();
        nioAuthAcceptor.setDefaultLocalAddress(new InetSocketAddress(authport));
        DatagramSessionConfig dcfg = nioAuthAcceptor.getSessionConfig();
        dcfg.setReceiveBufferSize(33554432);
        dcfg.setReadBufferSize(8192);
        dcfg.setSendBufferSize(8192);
        dcfg.setBothIdleTime(0);
        dcfg.setReuseAddress(false);


        DefaultIoFilterChainBuilder authIoFilterChainBuilder = new DefaultIoFilterChainBuilder();
        ExecutorFilter authExecutorFilter = new ExecutorFilter(3, getAuthPool(), 60, TimeUnit.SECONDS);
        Map<String, IoFilter> filters = new LinkedHashMap<>();
        filters.put("executor", authExecutorFilter);
        authIoFilterChainBuilder.setFilters(filters);
        nioAuthAcceptor.setFilterChainBuilder(authIoFilterChainBuilder);
        nioAuthAcceptor.setHandler(radiusAuthHandler);
        nioAuthAcceptor.bind();
        logger.info(String.format("====== RadiusAuthServer listen %s ======", authport));
        return nioAuthAcceptor;
    }

    /**
     * Radius 记账服务配置
     * @param radiusAcctHandler
     * @return
     * @throws IOException
     */
    @Bean(destroyMethod = "unbind")
    public NioDatagramAcceptor nioAcctAcceptor(RadiusAcctHandler radiusAcctHandler) throws IOException {
        if(!running){
            logger.info("====== RadiusAcctServer not running ======");
            return null;
        }
        NioDatagramAcceptor nioAcctAcceptor = new NioDatagramAcceptor();
        nioAcctAcceptor.setDefaultLocalAddress(new InetSocketAddress(acctport));
        DatagramSessionConfig dcfg = nioAcctAcceptor.getSessionConfig();
        dcfg.setReceiveBufferSize(33554432);
        dcfg.setReadBufferSize(8192);
        dcfg.setSendBufferSize(8192);
        dcfg.setBothIdleTime(0);
        dcfg.setReuseAddress(false);

        DefaultIoFilterChainBuilder acctIoFilterChainBuilder = new DefaultIoFilterChainBuilder();
        ExecutorFilter acctExecutorFilter = new ExecutorFilter(3, getAcctPool(), 60, TimeUnit.SECONDS);
        Map<String, IoFilter> filters = new LinkedHashMap<>();
        filters.put("executor", acctExecutorFilter);
        acctIoFilterChainBuilder.setFilters(filters);
        nioAcctAcceptor.setFilterChainBuilder(acctIoFilterChainBuilder);
        nioAcctAcceptor.setHandler(radiusAcctHandler);
        nioAcctAcceptor.bind();
        logger.info(String.format("====== RadiusAcctServer listen %s ======", acctport));
        return nioAcctAcceptor;
    }


    public int getAuthport() {
        return authport;
    }

    public void setAuthport(int authport) {
        this.authport = authport;
    }

    public int getEventport() {
        return eventport;
    }

    public void setEventport(int eventport) {
        this.eventport = eventport;
    }

    public int getAcctport() {
        return acctport;
    }

    public void setAcctport(int acctport) {
        this.acctport = acctport;
    }

    public int getRejectdelayTimes() {
        return rejectdelayTimes;
    }

    public void setRejectdelayTimes(int rejectdelayTimes) {
        this.rejectdelayTimes = rejectdelayTimes;
    }

    public int getTrace() {
        return trace;
    }

    public void setTrace(int trace) {
        this.trace = trace;
    }

    public int getInterimUpdate() {
        return interimUpdate;
    }

    public void setInterimUpdate(int interimUpdate) {
        this.interimUpdate = interimUpdate;
    }

    public int getMaxSessionTimeout() {
        return maxSessionTimeout;
    }

    public void setMaxSessionTimeout(int maxSessionTimeout) {
        this.maxSessionTimeout = maxSessionTimeout;
    }

    public String getTicketDir() {
        return ticketDir;
    }

    public void setTicketDir(String ticketDir) {
        this.ticketDir = ticketDir;
    }

    public boolean isRunning() {
        return running;
    }

    public void setRunning(boolean running) {
        this.running = running;
    }

    public boolean isBillInput() {
        return isBillInput;
    }

    public void setBillInput(boolean billInput) {
        isBillInput = billInput;
    }

    public boolean isBillBackFlow() {
        return isBillBackFlow;
    }

    public void setBillBackFlow(boolean billBackFlow) {
        isBillBackFlow = billBackFlow;
    }

    public boolean isAllowNegative() {
        return allowNegative;
    }

    public void setAllowNegative(boolean allowNegative) {
        this.allowNegative = allowNegative;
    }

    public int getRejectdelay() {
        return rejectdelay;
    }

    public void setRejectdelay(int rejectdelay) {
        this.rejectdelay = rejectdelay;
    }

    public int getTicketExpireDays() {
        return ticketExpireDays;
    }

    public void setTicketExpireDays(int ticketExpireDays) {
        this.ticketExpireDays = ticketExpireDays;
    }

    public boolean isTraceEnabled(){
        return trace == 1;
    }

    public int getAuthPool() {
        return authPool;
    }

    public void setAuthPool(int authPool) {
        this.authPool = authPool;
    }

    public int getAcctPool() {
        return acctPool;
    }

    public void setAcctPool(int acctPool) {
        this.acctPool = acctPool;
    }

    public String getStatfile() {
        return statfile;
    }

    public void setStatfile(String statfile) {
        this.statfile = statfile;
    }

    public int getRejectdelayEnabled() {
        return rejectdelayEnabled;
    }

    public void setRejectdelayEnabled(int rejectdelayEnabled) {
        this.rejectdelayEnabled = rejectdelayEnabled;
    }

    public String getStatDir() {
        return statDir;
    }

    public void setStatDir(String statDir) {
        this.statDir = statDir;
    }
}
