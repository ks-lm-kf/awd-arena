FROM python:3.11-alpine
RUN apk add --no-cache socat
RUN echo "flag{buff3r_0v3rfl0w_pwn_m4st3r_2026}" > /flag.txt && chmod 644 /flag.txt
COPY pwn.py /pwn.py
RUN chmod +x /pwn.py
EXPOSE 9999
CMD ["socat", "TCP-LISTEN:9999,reuseaddr,fork", "EXEC:python3 /pwn.py"]

