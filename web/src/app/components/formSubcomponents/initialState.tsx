// non-client even though it uses state?

import {Note} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";

export function InitialNotesState(existingNotes?: Note[]): AllEntries<Note> {
    return {
        // This was dataFor
        existing: (existingNotes || []).map((l) => {
            return {data: l, disabled: false}
        }),
        new: []
    }
}