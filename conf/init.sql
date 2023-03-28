--  This file is part of the eliona project.
--  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
--  ______ _ _
-- |  ____| (_)
-- | |__  | |_  ___  _ __   __ _
-- |  __| | | |/ _ \| '_ \ / _` |
-- | |____| | | (_) | | | | (_| |
-- |______|_|_|\___/|_| |_|\__,_|
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
--  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
--  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
--  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

create schema if not exists glutz;

create table if not exists glutz.config
(
    config_id           bigserial primary key,
    username            text not null,
    password            text not null,
    api_token           text not null,
    url                 text not null,
    active              boolean default false,
    enable              boolean default false,
    request_timeout     integer default 120,
    refresh_interval    integer default 60,
    default_openable_duration   integer default 10,
    initialized      boolean default false,
    project_ids          text[]
);

create table if not exists glutz.devices
(
    config_id           bigint not null,
    project_id          text not null,
    device_id           text not null,
    asset_id            integer not null,
    location_id         text not null,
    primary key(config_id, project_id, device_id)
);


commit;

