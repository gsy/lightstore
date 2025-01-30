package com.lightstore.user.domain.shared;

import java.io.Serializable;

public interface Entity<E, ID extends Serializable> {
    boolean sameIdentityAs(E other);
    ID id();
}
