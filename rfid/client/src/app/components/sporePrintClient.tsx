'use client'

import React, {useState} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    InlineSubArea,
    InlineProps,
    OptionalArrayOfType,
    OptionalSimpleKey,
    RequiredKey,
    DisplayInput,
    InlineExpansionArea,
    ImportDisplayInput,
    resolvePicsFormData,
    setFormImages,
    OptionalKey,
    HandleTxtResponse,
    HandleJsonResponse,
    SendMultipartRequest,
    setFormData, InlineExpansionButton
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    MostRecentImageDisplay,
    PicsDisplay, SpeciesArea, SporePrintColorArea, SporePrintDensityArea, SubspeciesArea,
} from "@/app/components/formSubcomponents/commonClient";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {
    IsValidNote, NewEntryNotes,
    Note,
    NotesAreaInline
} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {SaleArea} from "@/app/components/saleClient";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {BaseExternalUrl} from "@/app/components/Constants";
import {redirect} from "next/navigation";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {NewMssForm, RecentSelectorV2} from "@/app/components/mssClient";
import {FruitData} from "@/app/components/fruitServer";
import {dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {NewSporeSwabForm} from "@/app/components/sporeSwabClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/pcRunClient";
import {FruitRecentSelector} from "@/app/components/fruitClient";
import {CreatedLinkFor} from "@/app/components/substrateRecipeClient";
import {MssData} from "@/app/components/mssServer";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {FlexedArea, FlexedSinglesGroup, NotesFormArea} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";

export function AssertSporePrint(input: any): asserts input is SporePrintData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['subspecies', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
        ['color', 'string'],
        ['density', 'string'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Plate assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['pics', IsValidPicWithNotesIncoming], // TODO: ensure ok
        ['notes', IsValidNote], // TODO: ensure ok
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function SporePrintImportDisplay({headerLevel, cookies}:ImportDisplayInput) { // TODO: USE ONLY FOR EXISTING SPORE PRINTS!
    const [printDate, setPrintDate] = useState<number>(Date.now())
    const [color, setColor] = useState<string | undefined>()
    const [density, setDensity] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [image, setImage] = useState<File | undefined>()
    const [err, setErr] = useState<string | undefined>()
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    const importEntry = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(!species){
            setErr("A species must be selected")
            return
        }
        let body = new FormData()
        let dataObj:any = {
            color: color, // TODO: validate on insert
            density: density, // TODO: validate on insert
            printDate:printDate,
            species:species._id,
            notes:notes,
            //perms: perms, // TODO: validate on insert
        }
        subspecies && (dataObj.subspecies = subspecies._id) // TODO: validate on in
        setFormData(body, dataObj)
        //body.set("data",JSON.stringify(dataObj))
        // Img
        if(image!==undefined){
            body.set("img",image,"img")
        }

        SendMultipartRequest( BaseExternalUrl + "db/import/sporePrint", cookies, body)
            .then(HandleTxtResponse)
            .then(id=>{
                redirect(BaseExternalUrl+"/view/sporePrint/"+id)
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }
    //no parent because we couldn't possibly know it
        return <ImportEntryFormWrapper entryType={"sporePrint"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <DateArea pre={"Print Date: "} readonly={false} when={Date.now()} updateParent={setPrintDate}/>
            <SporePrintColorArea readonly={false} setColor={setColor} />
            <SporePrintDensityArea readonly={false} setDensity={setColor} />
            <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
            <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
            <ImageSelector updateParent={setImage}/>
            <NewEntryNotes setNotes={setNotes} />
            {/*<EntryPermsArea setEntryPerms={setPerms}/>*/}
            <button className={"greenButton"} onClick={importEntry}>{"Create"}</button>
        </ImportEntryFormWrapper>

}

export default function SporePrintDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertSporePrint(data)
        const [initial, setInitial] = useState(data)
        // TODO: DO UPDATEINITIAL!!!
        const initNotes: Data<Note>[] = (data.notes || []).map((n)=>{
            return {data: n,disabled:false}
        })
        const [color, setColor] = useState<string | undefined>()
        const [density, setDensity] = useState<string | undefined>()
        const [pics, setPics] = useState(InitialPicsEntries(data.pics))
        const [sale, setSale] = useState(data.sale)
        const [disposed, setDisposed] = useState(data.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: initNotes,new:[]})
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(data.acl)
        const updateInitial= (updated: SporePrintData)=>{
            setInitial(updated)
            setColor(updated.color)
            setDensity(updated.density)
            setPics(InitialPicsEntries(updated.pics))
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const submit = ()=>{
            // sale disposed, project, pics, notes
            let body = new FormData()
            let dataObj:any={
                color: color,
                density: density,
                sale:sale,
                disposed:disposed,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                //body.set("data",JSON.stringify(dataObj)) // TODO: REDO THINGS ON GO SIDE (ENSURE OTHER PICTURE ONES DO THE SAME!
                setFormImages(body, "newPic", newImages)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl+"/db/update/sporePrint/"+data._id, cookies, body)
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSporePrint(entry)
                    updateInitial(entry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: test heavily for all
            // TODO: Chain spore print (do not allow after too long) ---------------------------- TODO!!!!
            { // TODO: create SporeSwab
                txt: "Create Spore Swab",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporeSwabForm printIn={data} onCreate={(item: MssData)=>{ // TODO: switch to handlers{{}} format
                        onCreate([{
                            typeText: "Multispore Syringe",
                            node: <CreatedLinkFor linkId={item._id} typ={"mss"}/>,
                        }])
                    }}/>
                }
            },
            { // TODO: create MSS
                txt: "Create MultiSpore Syringe",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewMssForm sporePrintIn={data} handlers={{
                        isTopLevel: false,
                        onCreate: (item: MssData)=>{
                            onCreate([{
                                typeText: "Multispore Syringe",
                                node: <CreatedLinkFor linkId={item._id} typ={"mss"}/>,
                            }])
                        }
                    }} />
                }
            },
            // TODO: any transfers ok???
            // TODO: this????OvcForXfers(data._id, "sporePrint", ["bag","fruitingChamber","jar","plate","slant","stasisTube"], AddToTransfers(setTransfersOut, transfersOut)), // TODO: ensure list correct
        ]
        return <DisplayFormWrapper entryType={"sporePrint"}>
            <ErrorDisplay err={err} headerLevel={headerLevel} />
            <ID id={data._id} txt={"Spore Print"} entryType={"sporePrint"}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: is this position ok???? */}{/* TODO: OR TRI??? where to put? Chain print, swab from print, print transfer to agar*/}
            <MostRecentImageDisplay data={data.mostRecentImage} headerLevel={headerLevel}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <DateArea pre={"Print Date: "} readonly={true} when={data.creationDate}/>
                    <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
                    <DisposedDisplay readonly={false} disposed={disposed} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <div>
                        <div>{"Parent: "}</div>
                        {data.parent?
                            <EntryLink props={{displayedId:data.parent,linkId:data.parent,entryType:"fruit",openInNewTab:true}}>{data.parent}</EntryLink>
                            :"Store"}
                    </div>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SporePrintColorArea readonly={true} color={data.color}/>
                    <SporePrintDensityArea readonly={true} density={data.density}/>
                    <SaleArea readonly={false} canCreateSale={true} sale={sale} setSale={setSale} headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            {/* TODO: area where we can display all the child MSS of this print? */}
            <PicsDisplay pix={pics} updateParent={setPics} readonly={readonly} headerLevel={headerLevel}/>{/* Pics */}
            <NotesFormArea initial={initial.notes} readonly={readonly} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Spore print data format incorrect: " + err}</div>
    }
}

// Should only be accessible from a fruit's page
export function NewSporePrintForm(
    {fruitIn, headerLevel, offset, onCreate, cookies}: { // TODO: isTopLevel or whatever
        fruitIn?: FruitData
        headerLevel?: number
        offset?: number
        onCreate:(sp: SporePrintData)=>void
        cookies: string
}){
    const [fruit, setFruit] = useState<FruitData | undefined>(fruitIn)
    const [pics, setPics] = useState<NewPicWithNotesForm[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    // Spore prints don't have rfid tags, although they have MainCollectionIDs
    const [err, setErr] = useState<string | undefined>(undefined)
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    const createEntry = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(!fruit){
            setErr("Fruit must be selected")
            return
        }
        // TODO: if both pics and notes are empty, do nothing
        if(pics.length===0){
            setErr("Must at least contain one picture")
            return
        }
        let body = new FormData()
        let dataObj:any = {
            notes:notes,
            fruitId:fruit._id,
            // No perms means inherit from parents?
        }
        // Pics
        dataObj.pics = pics.map(p=>{return {time:p.time,notes:p.notes.new.map(n => {
                return n.data
            })}})
        // Perms
        // if(perms!==undefined){
        //     dataObj.perms = perms
        // }
        setFormData(body, dataObj)
        //body.set("data", JSON.stringify(dataObj))
        for (let i = 0; i < pics.length; i++) {
            let toSend = pics[i]
            if (toSend.img === undefined) {
                setErr("new image " + i + " is undefined")
                return
            }
            const fileName = "newPic" + "-" + i
            body.set(fileName, toSend.img, fileName)
        }

        SendMultipartRequest(BaseExternalUrl+"/create/sporePrint", cookies, body)
            .then(HandleJsonResponse)
            .then((resJson)=>{
                AssertSporePrint(resJson) // TODO: make sure comes back as print obj
                onCreate(resJson)
            })
            .catch((er) => {
                setErr(JSON.stringify(er))
            });
    }
    if(fruitIn===undefined){
        // TODO: FRUIT SELECTOR!
    }

    return <NewEntryFormWrapper entryType={"sporePrint"}>
        <ErrorDisplay err={err} headerLevel={headerLevel} offset={offset}/>
        {fruitIn === undefined && <FruitRecentSelector onSelect={setFruit}/>}{/* TODO: isTopLevel stuff???*/}
        <PicsDisplay pix={{existing:[],new:dataFor(pics)}} readonly={false} updateParent={(ps)=>{setPics(ps.new.map((p)=>{return p.data}))}} headerLevel={headerLevel} offset={offset}/>
        <NewEntryNotes setNotes={setNotes} />
        {/*<EntryPermsArea setEntryPerms={setPerms}/> /!* TODO: only use perms if we want to? *!/*/}
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function SporePrintInline(
    {
        data, expandByDefault, headerLevel, onClick, showMainPageButton, idIsLink
    }: InlineProps<SporePrintData>
) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Spore Print"} entryType={"sporePrint"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <MostRecentImageDisplay data={data.mostRecentImage} headerLevel={headerLevel} />
            <DateArea pre={"Print Date: "} readonly={true} when={data.creationDate} />
            <SpeciesArea readonly={true} headerLevel={headerLevel} initial={data.species}/>
            <SubspeciesArea readonly={true} headerLevel={headerLevel} currentSpecies={data.species} initialSub={data.subspecies}/>
            <SaleArea readonly={true} canCreateSale={false} sale={data.sale} headerLevel={headerLevel}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded: expanded}}>
            <div>
                <div>{"Parent" + (data.parent ? <EntryLink props={{displayedId:data.parent,linkId:data.parent,entryType:"fruit",openInNewTab:true}}>{data.parent}</EntryLink> : "unknown")}</div>{/* TODO: FIX for link! */}
            </div>
            {/* TODO: parent?: string // Only empty if purchased and not printed yourself*/}
            {/*TODO: <ProjectsArea allowCreate={false} allowRemove={false} projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel} offset={-1}/>*/}
            <div>
                <div>{"Color" + (data.color || "unknown")}</div>
            </div>
            <div>
                <div>{"Density" + (data.density || "unknown")}</div>
            </div>
            <NotesAreaInline notes={data.notes} headerLevel={headerLevel} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// export function SporePrintListDisplay({data, onClick}: SingleListProps<SporePrintData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <SporePrintInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }

// TODO: HEAVILY TEST!!!!
export function SporePrintRecentSelector({onSelect}:{onSelect:(selected?: SporePrintData) => void}) {
    return <RecentSelectorV2<SporePrintData> listUrlType={"sporePrints"} assertion={AssertSporePrint} singleConstructor={(val, i)=>{
        return <SporePrintInline data={val} expandByDefault={false} onClick={onSelect}/>
    }} />
}
