package org.toughradius.component;

import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Bras;
import org.toughradius.mapper.BrasMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class BrasService {

	@Autowired
	private BrasMapper tcBrasMapper;

	@Autowired
	private Syslogger logger;

	public Bras findBras(String ipaddr, String srcip, String identifier) throws ServiceException{
		Bras tcBras = null;
		if(ValidateUtil.isNotEmpty(ipaddr)&&!"0.0.0.0".equals(ipaddr)){
			tcBras = tcBrasMapper.selectByIPAddr(ipaddr);
		}

		if(ValidateUtil.isNotEmpty(srcip)&&!"0.0.0.0".equals(srcip)){
			tcBras = tcBrasMapper.selectByIPAddr(srcip);
		}

		if (tcBras == null && ValidateUtil.isNotEmpty(identifier)) {
			tcBras = tcBrasMapper.selectByidentifier(identifier);
		}

		if (tcBras == null) {
			String message = String.format("Bras设备 id=%s, ip=%s 不存在", identifier, ipaddr);
			logger.error(message,Syslogger.RADIUSD);
			throw new ServiceException(message);
		}

		if (tcBras.getStatus() != null && "disabled".equals(tcBras.getStatus())) {
			String message = String.format("Bras设备 id=%s, ip=%s 已停用", identifier, ipaddr);
			logger.error(message,Syslogger.RADIUSD);
			throw new ServiceException(message);
		}

		return tcBras;
	}
}
