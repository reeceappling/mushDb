import {NewPicWithNotesForm} from "@/app/components/formSubcomponents/picWithNotes";
import {useState} from "react";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import DateArea from "@/app/components/formSubcomponents/date";
import NotesArea, {NotesAreaOld, Note, NotesGrid} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";

export function PixRowNew( // TODO: THIS!
    {current, updateParent, remv}: {
        current: NewPicWithNotesForm
        updateParent?: (d: NewPicWithNotesForm) => void,
        remv: () => void,
    }){
    const [editing, setEditing] = useState(true)
    const leftArea = () => {
        return <div className={"picLeft"}>
            {/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */}
            <ImageSelector updateParent={f => {
                let upd = {...current}
                upd.img = f
                updateParent && updateParent(upd)
            }}/>
            <button className={"removeButton"} onClick={remv}>{"REMOVE THIS Entry"}</button>
        </div>
    }
    const rightArea = () => {
        return <div className={"picRight"}>
            <DateArea readonly={true} when={current.time}/>


            {/* TODO: CLASSES!!! */}
            {/* TODO: UNMODIFIABLE DATES ON NEW NOTES */}
            {/* TODO: try notes grid!*/}
            <NotesGrid readonly={false} current={current.notes} updateParent={(nts: AllEntries<Note>) => {
                    let out = {...current} // TODO: notesFormArea?
                    out.notes = nts
                    updateParent && updateParent(out)
                }}/>
            {/*<NotesAreaOld readonly={false} current={current.notes} updateParent={(nts: AllEntries<Note>) => {*/}
            {/*        let out = {...current} // TODO: notesFormArea?*/}
            {/*        out.notes = nts*/}
            {/*        updateParent && updateParent(out)*/}
            {/*    }}/>*/}
            {<button className={"basicButton"} onClick={() => {setEditing(!editing)}}>
                {(editing ? "Save" : "Edit") + " Picture Notes"}
            </button>}
        </div>
    }
    return <div className={"picRow "}>
        {leftArea()}
        {rightArea()}
    </div>
}