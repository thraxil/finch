FROM scratch
COPY finch /
COPY media /media
COPY templates /templates
ENV FINCH_PORT 8000
ENV FINCH_DB_FILE /database.db
ENV FINCH_MEDIA_DIR /media
ENV FINCH_TEMPLATE_DIR /templates
ENV FINCH_ITEMS_PER_PAGE 50
EXPOSE 8000
CMD ["/finch"]
