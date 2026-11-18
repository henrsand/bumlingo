--
-- PostgreSQL database dump
--

\restrict BNc5vS1hQdOWQb0I5L010TtXbcTtBxwjS7YDrQjNfOueGWm5erhSFnHYdgjc0DX

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: poi; Type: TABLE; Schema: public; Owner: henrik_gis
--

CREATE TABLE public.poi (
    id bigint NOT NULL,
    label text NOT NULL,
    geom public.geometry(Point,4326) NOT NULL
);


ALTER TABLE public.poi OWNER TO henrik_gis;

--
-- Name: poi_id_seq; Type: SEQUENCE; Schema: public; Owner: henrik_gis
--

ALTER TABLE public.poi ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.poi_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: poi poi_pkey; Type: CONSTRAINT; Schema: public; Owner: henrik_gis
--

ALTER TABLE ONLY public.poi
    ADD CONSTRAINT poi_pkey PRIMARY KEY (id);


--
-- Name: poi_geom_idx; Type: INDEX; Schema: public; Owner: henrik_gis
--

CREATE INDEX poi_geom_idx ON public.poi USING gist (geom);


--
-- PostgreSQL database dump complete
--

\unrestrict BNc5vS1hQdOWQb0I5L010TtXbcTtBxwjS7YDrQjNfOueGWm5erhSFnHYdgjc0DX

