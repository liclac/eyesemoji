pub mod errors;

use blurz::{
    BluetoothAdapter, BluetoothDevice, BluetoothDiscoverySession,
    BluetoothGATTCharacteristic as GATTCharacteristic, BluetoothGATTDescriptor as GATTDescriptor,
    BluetoothGATTService as GATTService, BluetoothSession,
};
use std::thread;
use std::time::Duration;

fn warn_or_else<T, E: std::fmt::Display, F: FnOnce() -> T>(result: Result<T, E>, default: F) -> T {
    match result {
        Ok(v) => v,
        Err(e) => {
            println!("[WARN] {:}", e);
            default()
        }
    }
}

fn warn_or<T, E: std::fmt::Display>(result: Result<T, E>, default: T) -> T {
    warn_or_else(result, move || default)
}

fn warn<T: Default, E: std::fmt::Display>(result: Result<T, E>) -> T {
    warn_or_else(result, T::default)
}

fn match_device(dev: &BluetoothDevice) -> Result<Option<String>, Box<dyn ::std::error::Error>> {
    println!(
        "[{:}] class={:#X} icon={:} look={:#X} alias={:} name={:}",
        dev.get_address().unwrap_or("00:00:00:00:00:00".to_string()),
        dev.get_class().unwrap_or(0),
        dev.get_icon().unwrap_or("".to_string()),
        dev.get_appearance().unwrap_or(0),
        dev.get_alias().unwrap_or("".to_string()),
        dev.get_name().unwrap_or("".to_string()),
    );
    println!(
        "  vendor={:#X} product={:#X} device={:#X}",
        dev.get_vendor_id().unwrap_or(0),
        dev.get_product_id().unwrap_or(0),
        dev.get_device_id().unwrap_or(0),
    );
    println!("  manufacturer_data={:?}", dev.get_manufacturer_data().ok());
    println!("  service_data={:?}", dev.get_service_data().ok());
    println!("  gatt_services={:?}", dev.get_gatt_services().ok());

    println!("  uuids=");
    let mut hit = None;
    for uuid in dev.get_uuids()? {
        println!("    - {:}", uuid);
        if uuid == "0000fff0-0000-1000-8000-00805f9b34fb" {
            println!("      ^^  HIT!  ^^^^^^^^^^^^^^^^^^^^^^^^^^");
            hit = Some(uuid.to_string());
        }
    }
    Ok(hit)
}

fn main() -> Result<(), Box<dyn ::std::error::Error>> {
    // Create a Bluetooth session, and an Adapter (eg. daemon connection).
    let sess = BluetoothSession::create_session(None)?;
    let adapter = BluetoothAdapter::init(&sess)?;

    // Start scanning for anything we can find.
    let disc_sess = BluetoothDiscoverySession::create_session(&sess, adapter.get_id())?;
    disc_sess.start_discovery()?;

    // Find a device we recognise.
    let mut dev = BluetoothDevice::new(&sess, "".into());
    let mut profile = "".to_string();
    while profile == "" {
        println!("----- BEGIN -----");
        for path in adapter.get_device_list()? {
            let candidate = BluetoothDevice::new(&sess, path);
            match match_device(&candidate) {
                Ok(ouuid) => {
                    if let Some(uuid) = ouuid {
                        dev = candidate;
                        profile = uuid;
                    }
                }
                Err(e) => {
                    println!("[{:}] {:}", candidate.get_id(), e);
                }
            }
        }
        println!("-----  END  -----");
        thread::sleep(Duration::from_millis(2000));
    }

    println!("We found a hit, stopping discovery...");
    disc_sess.stop_discovery()?;

    // Hey, come here often?
    println!(
        "Connecting to: {:} ({:})...",
        dev.get_id(),
        dev.get_alias().unwrap_or("".to_string())
    );
    dev.connect(1000)?;
    println!("Connected!");

    // Whatchu got?
    println!("Listing GATT services...");
    let mut svcs;
    loop {
        svcs = dev.get_gatt_services()?;
        if !svcs.is_empty() {
            break;
        }
        println!("Waiting for GATT services to appear...");
        thread::sleep(Duration::from_millis(1000));
    }

    // These glasses only have a single service.
    let svc = GATTService::new(&sess, svcs.first().unwrap().to_string());
    println!("{:}", svc.get_id());
    println!("  uuid={:}", warn(svc.get_uuid()));
    println!("  primary={:?}", warn(svc.is_primary()));
    println!("  includes={:?}", warn(svc.get_includes()));
    println!("  characteristics=");
    for chid in svc.get_gatt_characteristics()? {
        let ch = GATTCharacteristic::new(&sess, chid);
        println!("  - {:}", ch.get_id());
        println!("    uuid={:}", warn(ch.get_uuid()));
        println!("    flags={:?}", warn(ch.get_flags()));
        println!("    value={:#X?}", warn(ch.get_value()));
        println!("    descriptors=");
        for did in ch.get_gatt_descriptors()? {
            let desc = GATTDescriptor::new(&sess, did);
            println!("    - {:}", desc.get_id());
            println!("      uuid={:}", warn(desc.get_uuid()));
            println!("      flags={:?}", warn(desc.get_flags()));
            println!("      value={:#X?}", warn(desc.get_value()));
        }
        loop {
            // Turn them off...
            println!("Turning them off...");
            ch.write_value(vec![0x01, 0x00, 0x02, 0x06, 0x09, 0x00, 0x03], None)?;
            thread::sleep(Duration::from_millis(1000));

            // ...and on again!
            println!("  ...and on again!");
            ch.write_value(vec![0x01, 0x00, 0x02, 0x06, 0x09, 0x02, 0x05, 0x03], None)?;
            thread::sleep(Duration::from_millis(1000));
        }
    }

    Ok(())
}
