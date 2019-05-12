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

    void deleteById(Long id);

    Bras findByidentifier(String identifier);

    Bras findByIPAddr(String ipaddr);

    Bras findById(Long id);

}