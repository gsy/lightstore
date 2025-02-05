package com.lightstore.user.infrastructure.persistence.repository;

import com.lightstore.user.domain.entity.User;
import com.lightstore.user.domain.entity.UserRepository;
import com.lightstore.user.infrastructure.persistence.mapping.UserModel;
import org.springframework.stereotype.Repository;


@Repository
public class UserRepositoryImpl implements UserRepository {

    private final JdbcUserRepository impl;

    public UserRepositoryImpl(JdbcUserRepository repository) {
        this.impl = repository;
    }

    @Override
    public User save(User user) {
        final UserModel model = new UserModel();
        model.setUserID(user.id().id);
        model.setUsername(user.getUsername());
        UserModel saved = impl.save(model);
        return user;
    }
}
