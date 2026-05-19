use smallvec::SmallVec;
use std::fmt::Debug;

use tracing::field::Field;

#[derive(Default)]
pub(super) struct EventVisitor {
    pub(crate) message: Option<String>,
    pub(crate) fields: SmallVec<[(&'static str, String); 4]>,
}

impl EventVisitor {
    pub(super) fn record_field(&mut self, name: &'static str, value: String) {
        if name == "msg" || name == "message" {
            self.message = Some(value);
            return;
        }
        self.fields.push((name, value));
    }
}

impl tracing::field::Visit for EventVisitor {
    fn record_i64(&mut self, field: &Field, value: i64) {
        self.record_field(field.name(), value.to_string());
    }

    fn record_u64(&mut self, field: &Field, value: u64) {
        self.record_field(field.name(), value.to_string());
    }

    fn record_bool(&mut self, field: &Field, value: bool) {
        self.record_field(field.name(), value.to_string());
    }

    fn record_str(&mut self, field: &Field, value: &str) {
        self.record_field(field.name(), value.to_owned());
    }

    fn record_debug(&mut self, field: &Field, value: &dyn Debug) {
        self.record_field(field.name(), format!("{value:?}"));
    }
}
