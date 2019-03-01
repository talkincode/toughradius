package org.toughradius.mapper;

import org.toughradius.entity.Bras;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

@Repository
@Mapper
public interface BrasMapper {
	

    Bras selectByIPAddr(String ipaddr);
    
    Bras selectByidentifier(String identifier);

    int  flushCache();
}