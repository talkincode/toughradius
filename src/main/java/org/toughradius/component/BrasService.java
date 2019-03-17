package org.toughradius.component;

import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Bras;
import org.toughradius.mapper.BrasMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class BrasService {

	@Autowired
	private BrasMapper brasMapper;

	@Autowired
	private Syslogger logger;

	public Bras findBras(String ipaddr, String srcip, String identifier) throws ServiceException{
		Bras tcBras = null;
		if(ValidateUtil.isNotEmpty(ipaddr)&&!"0.0.0.0".equals(ipaddr)){
			tcBras = brasMapper.selectByIPAddr(ipaddr);
		}

		if(ValidateUtil.isNotEmpty(srcip)&&!"0.0.0.0".equals(srcip)){
			tcBras = brasMapper.selectByIPAddr(srcip);
		}

		if (tcBras == null && ValidateUtil.isNotEmpty(identifier)) {
			tcBras = brasMapper.selectByidentifier(identifier);
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

	public List<Bras> queryForList(Bras bras){
		return brasMapper.queryForList(bras);
	}

	public void insertBras(Bras bras){
		brasMapper.insertBras(bras);
	}

	public void updateBras(Bras bras){
		brasMapper.updateBras(bras);
	}

	public void deleteById(Integer id){
		brasMapper.deleteById(id);
	}

	public Bras selectByidentifier(String identifier){
		return brasMapper.selectByidentifier(identifier);
	}

	public Bras selectByIPAddr(String ipaddr){
		return brasMapper.selectByIPAddr(ipaddr);
	}

	public Bras selectById(Integer id){
		return brasMapper.selectById(id);
	}
}
