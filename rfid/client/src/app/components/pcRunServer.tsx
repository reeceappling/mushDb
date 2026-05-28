import {Note} from "@/app/components/formSubcomponents/notes";
import {NewPcRunForm, PcRunSelector} from "@/app/components/pcRunClient";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ACL} from "@/app/components/accessControlServer";

export function TestPcRunOk(){
    const a: PcRunData = {
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
    }
    return a
}

export interface PcRunData {
    _id: string
    creationDate: number
    runtimeMinutes: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

// TODO: VALIDATE WORKS!
export function PcRunSelectorCloseable(sp: SelectorProps<PcRunData>){ // TODO: likely overhaul
    return <CloseableSelector<PcRunData> props={{
        allowCreation: sp.allowCreation,
        doSelect: sp.doSelect, // For selecting normally
        msgTxt: ChannelTextNewPcRun, // TODO: del?
        closeTxt: "Close PcRun List",
        createTxt: "Create Pc Run",
        createEndpt: "pcRun",
        lowercase: "pc run",
        creatorInPage: sp.creatorInPage,
        getId: (v: PcRunData)=>v._id,
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