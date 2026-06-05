import {Note} from "@/app/components/formSubcomponents/notes";
import {NewPcRunForm, PcRunSelector} from "@/app/components/pcRunClient";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ACL, TestAcl} from "@/app/components/accessControlServer";

export function TestPcRunOk(){
    return new PcRunData({
        _id: "(ID_HERE)",
        creationDate: Date.now()-2000,
        runtimeMinutes: 120,
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        },{
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: Date.now(),
        acl: TestAcl(), // TODO: do we want?
    })
}

export interface PcRunData {
    _id: string
    creationDate: number
    runtimeMinutes: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class PcRunData {
    // Accept a single object containing the fields
    constructor(init?: Partial<PcRunData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "pcRun"
    }
}

// TODO: VALIDATE WORKS!
export function PcRunSelectorCloseable(sp: SelectorProps<PcRunData>){ // TODO: likely overhaul
    return <CloseableSelector<PcRunData> props={{
        allowCreation: sp.allowCreation,
        doSelect: sp.doSelect, // For selecting normally // TODO: ALLOW DESELECT/CLEAR!
        closeTxt: "Close PcRun List",
        createTxt: "Create Pc Run",
        createEndpt: "pcRun",
        lowercase: "pc run",
        creatorInPage: sp.creatorInPage,
        createSelector:(selHdl: (onSelect: PcRunData) => void)=>{
            return <PcRunSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: PcRunData) => void)=>{
            return <NewPcRunForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}

const ChannelTextNewPcRun = "newPcRun"