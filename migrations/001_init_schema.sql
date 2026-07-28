-- Миграция для создания схемы базы данных Postgres SQL ч 1

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
	    id SERIAL PRIMARY KEY,
	    username VARCHAR(50) NOT NULL UNIQUE,
	    email VARCHAR(255) NOT NULL UNIQUE CHECK (email LIKE '%_@__%.__%'),
	    password VARCHAR(255) NOT NULL CHECK (LENGTH(password) >= 8),
	    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица постов бзера
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL CHECK (LENGTH(title) > 0),
    content TEXT NOT NULL	CHECK (LENGTH(content) > 0),
    author_id INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',	
    publish_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT fk_posts_author 
        	FOREIGN KEY (author_id) 
	        REFERENCES users(id) 
     		ON DELETE CASCADE
);

-- Таблица комментариев к посту юзера
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL CHECK (LENGTH(content) > 0),
    post_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT fk_comments_post 
		FOREIGN KEY (post_id) 
		REFERENCES posts(id) 
		ON DELETE CASCADE,

	CONSTRAINT fk_comments_author 
		FOREIGN KEY (author_id) 
		REFERENCES users(id) 
		ON DELETE CASCADE
);
