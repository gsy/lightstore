package com.lightstore.user.infrastructure.persistence.repository;

import com.lightstore.user.domain.entity.User;
import com.lightstore.user.domain.entity.UserRepository;
import com.lightstore.user.infrastructure.persistence.mapping.UserTable;
import org.springframework.stereotype.Repository;


@Repository
public class UserRepositoryImpl implements UserRepository {

    private final JdbcUserRepository impl;

    public UserRepositoryImpl(JdbcUserRepository repository) {
        this.impl = repository;
    }

    @Override
    public User save(User user) {
        final UserTable obj = new UserTable();
        obj.setUserID(user.id().id);
        obj.setUsername(user.getUsername());
        UserTable saved = impl.save(obj);
        return user;
    }
}
