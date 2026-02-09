CREATE SCHEMA IF NOT EXISTS `college_rag`
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_0900_ai_ci;

create table datasets_permission
(
    id           varchar(36)                         not null
        primary key,
    dataset_id   varchar(36)                         not null,
    teacher_id   varchar(36)                         not null,
    teacher_name varchar(60)                         not null,
    granted_by   varchar(60)                         not null,
    granted_at   timestamp default CURRENT_TIMESTAMP not null,
    constraint unique_topic_teacher
        unique (dataset_id, teacher_id)
)
    charset = utf8mb4;

create index idx_topic_id
    on datasets_permission (dataset_id);

create table topics
(
    id            varchar(36)                         not null
        primary key,
    title         varchar(255)                        not null,
    description   text                                null,
    created_by    varchar(50)                         not null,
    created_by_id varchar(36)                         null,
    created_at    timestamp default CURRENT_TIMESTAMP not null,
    updated_at    timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
    charset = utf8mb4;

create table topic_assignments
(
    id             varchar(36)                         not null
        primary key,
    topic_id       varchar(36)                         not null,
    student_id     varchar(50)                         not null,
    student_name   varchar(255)                        not null,
    assigned_by    varchar(50)                         not null,
    assigned_by_id varchar(36)                         null,
    assigned_at    timestamp default CURRENT_TIMESTAMP not null,
    constraint unique_topic_student
        unique (topic_id, student_id),
    constraint topic_assignments_ibfk_1
        foreign key (topic_id) references topics (id)
            on delete cascade
)
    charset = utf8mb4;

create table datasets
(
    id            varchar(36)                         not null
        primary key,
    user_id       varchar(255)                        not null,
    author        varchar(255)                        null,
    title         varchar(255)                        not null,
    file_path     varchar(500)                        not null,
    created_at    timestamp default CURRENT_TIMESTAMP null,
    updated_at    timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP,
    indexed_at    timestamp                           null,
    topic_id      varchar(36)                         null,
    assignment_id varchar(36)                         null,
    constraint fk_dataset_assignment
        foreign key (assignment_id) references topic_assignments (id)
            on delete set null,
    constraint fk_dataset_topic
        foreign key (topic_id) references topics (id)
            on delete set null
)
    charset = utf8mb4;

create table dataset_analytics
(
    id          varchar(36)                                       not null
        primary key,
    dataset_id  varchar(36)                                       not null,
    user_id     varchar(255)                                      not null,
    action_type enum ('view', 'download', 'ask_question', 'edit') not null,
    question    text                                              null,
    created_at  timestamp default CURRENT_TIMESTAMP               null,
    constraint dataset_analytics_ibfk_1
        foreign key (dataset_id) references datasets (id)
            on delete cascade
)
    charset = utf8mb4;

create index idx_action_type
    on dataset_analytics (action_type);

create index idx_created_at
    on dataset_analytics (created_at);

create index idx_dataset_analytics
    on dataset_analytics (dataset_id);

create index idx_user_analytics
    on dataset_analytics (user_id);

create index idx_assignment_id
    on datasets (assignment_id);

create index idx_created_at
    on datasets (created_at);

create index idx_topic_id
    on datasets (topic_id);

create index idx_user_id
    on datasets (user_id);

create table saved_chats
(
    id         varchar(36)                         not null
        primary key,
    dataset_id varchar(36)                         not null,
    title      varchar(255)                        not null,
    created_by varchar(255)                        not null,
    user_id    varchar(36)                         not null,
    created_at timestamp default CURRENT_TIMESTAMP null,
    updated_at timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP,
    constraint fk_saved_chat_dataset
        foreign key (dataset_id) references datasets (id)
            on delete cascade
)
    charset = utf8mb4;

create table chat_messages
(
    id         varchar(36)                         not null
        primary key,
    chat_id    varchar(36)                         not null,
    question   text                                not null,
    answer     text                                not null,
    citations  json                                null,
    order_num  int                                 not null,
    created_at timestamp default CURRENT_TIMESTAMP null,
    constraint fk_chat_message_chat
        foreign key (chat_id) references saved_chats (id)
            on delete cascade
)
    charset = utf8mb4;

create index idx_chat_messages_chat_id
    on chat_messages (chat_id);

create index idx_saved_chats_dataset_id
    on saved_chats (dataset_id);

create index idx_saved_chats_user_id
    on saved_chats (user_id);

create index idx_student_id
    on topic_assignments (student_id);

create index idx_topic_id
    on topic_assignments (topic_id);

create index idx_created_at
    on topics (created_at);

create index idx_created_by
    on topics (created_by);

create index idx_created_by_id
    on topics (created_by_id);

