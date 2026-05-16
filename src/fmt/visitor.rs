use smallvec::SmallVec;
use std::fmt::Debug;

use tracing::field::Field;

#[derive(Default)]
pub(super) struct EventVisitor {
    pub(crate) message: Option<String>,
    pub(crate) fields: SmallVec<[(String, String); 4]>,
}

impl EventVisitor {
    pub(super) fn record_field(&mut self, name: &str, value: String) {
        if name == "message" || name == "msg" {
            self.message = Some(value);
        } else {
            self.fields.push((name.to_owned(), value));
        }
    }
}

impl tracing::field::Visit for EventVisitor {
    fn record_str(&mut self, field: &Field, value: &str) {
        self.record_field(field.name(), value.to_owned());
    }

    fn record_debug(&mut self, field: &Field, value: &dyn Debug) {
        self.record_field(field.name(), format!("{value:?}"));
    }
}
