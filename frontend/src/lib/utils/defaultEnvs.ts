export const DEFAULT_ENV: Record<string, Record<string, string>> = {
  mysql: {
    MYSQL_ROOT_PASSWORD: "vulnkit",
    MYSQL_DATABASE: "testdb",
  },
  mariadb: {
    MYSQL_ROOT_PASSWORD: "vulnkit",
    MYSQL_DATABASE: "testdb",
  },
  postgres: {
    POSTGRES_PASSWORD: "vulnkit",
    POSTGRES_DB: "testdb",
  },
  mongo: {
    MONGO_INITDB_ROOT_USERNAME: "root",
    MONGO_INITDB_ROOT_PASSWORD: "vulnkit",
  },
  redis: {},
  nginx: {},
  httpd: {},
};
