use log;
use serde::{Deserialize, Serialize};
use simplelog as sl;
use std::fs::{self, File};
use std::io::{Read, Write};
use std::net::{SocketAddr, TcpStream};
use std::str::from_utf8;
use std::thread;
use std::time::Duration;

#[derive(Debug, Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct Config {
    server: SocketAddr,
    initial_delay: Duration,
}

const CONFIG_FILE_NAME: &str = "config.json";

fn main() -> Result<(), std::io::Error> {
    sl::CombinedLogger::init(vec![
        sl::TermLogger::new(
            sl::LevelFilter::Trace,
            sl::Config::default(),
            sl::TerminalMode::Mixed,
            sl::ColorChoice::Auto,
        ),
        sl::WriteLogger::new(
            sl::LevelFilter::Trace,
            sl::Config::default(),
            File::create("log.txt").unwrap(),
        ),
    ])
    .unwrap();

    log::info!("Starting the client");

    if fs::metadata(CONFIG_FILE_NAME).is_err() {
        log::error!("Configuration file not found, generating a default one...");
        let config = Config {
            server: "127.0.0.1:7890".parse().unwrap(),
            initial_delay: Duration::from_secs(6),
        };
        serde_json::to_writer_pretty(File::create(CONFIG_FILE_NAME)?, &config)?;
        return Ok(());
    }
    let config: Config = {
        let config_file = File::open(CONFIG_FILE_NAME)?;
        serde_json::from_reader(&config_file)?
    };

    log::info!("Read config: {:?}", &config);

    match TcpStream::connect(&config.server) {
        Ok(mut stream) => {
            const STRING_TO_SEND: &str = "Романов Денис Игоревич";
            log::info!("Connected to {}", &config.server);
            log::info!("Waiting {:?} to send the string", &config.initial_delay);
            thread::sleep(config.initial_delay);

            log::info!("Sending {}", &STRING_TO_SEND);
            let mut send_buffer = Vec::new();
            std::write!(&mut send_buffer, "{}", STRING_TO_SEND).unwrap();

            let n = stream
                .write(&send_buffer)
                .expect("Couldn't write to server!");

            if let Ok(Some(e)) = stream.take_error() {
                return Err(e);
            }

            log::info!("Wrote {} bytes", n);

            // receiving
            let mut recv_buffer = [0 as u8; 1024];
            let n = stream
                .read(&mut recv_buffer)
                .expect("Couldn't read from server!");

            if let Ok(Some(e)) = stream.take_error() {
                return Err(e);
            }

            log::info!(
                "Received {} bytes from server: {}",
                n,
                from_utf8(&recv_buffer[..n]).unwrap()
            );
        }
        Err(_) => todo!(),
    }

    Ok(())
}
