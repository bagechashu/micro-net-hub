# Micro-Net-Hub Backend

Base on [go-ldap-admin-ui](https://github.com/eryajf/go-ldap-admin-ui), but **change too much** according to my personal habits.

# Comment

The page routes of the frontend project are regenerated by fetching data from the backend.

```
frontend/src/store/modules/permission.js > getRoutesFromMenuTree
```

# FIXME

- Currently, in order to add a `profile` button in the side navigation, it causes a conflict with the profile in `frontend/src/router/index.js`.
  > [vue-router] Duplicate named routes definition: { name: "Profile", path: "/profile/index" }
