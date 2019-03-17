package org.toughradius.mapper;

import org.toughradius.entity.Bras;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
@Mapper
public interface BrasMapper {



    List<Bras> queryForList(Bras bras);

    void insertBras(Bras bras);

    void updateBras(Bras bras);

    void deleteById(Integer id);

    Bras selectByidentifier(String identifier);

    Bras selectByIPAddr(String ipaddr);

    Bras selectById(Integer id);

    int  flushCache();
}